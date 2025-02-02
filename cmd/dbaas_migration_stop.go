package cmd

import (
	"context"
	"errors"
	"fmt"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type dbaasMigrationStopCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"stop"`

	Name string `cli-arg:"#"`
	Zone string `cli-short:"z" cli-usage:"Database Service zone"`
}

func (c *dbaasMigrationStopCmd) cmdAliases() []string { return []string{} }

func (c *dbaasMigrationStopCmd) cmdShort() string {
	return "Stop running migration of a Database"
}

func (c *dbaasMigrationStopCmd) cmdLong() string {
	return "This command stops the currently running migration of a Database."
}

func (c *dbaasMigrationStopCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasMigrationStopCmd) cmdRun(cmd *cobra.Command, args []string) error {
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	dbType, err := dbaasGetType(ctx, c.Name, c.Zone)
	if err != nil {
		if errors.Is(err, exoapi.ErrNotFound) {
			return fmt.Errorf("resource not found in zone %q", c.Zone)
		}
		return err
	}

	var stopMigrationFuncs = map[string]func(context.Context, string, string) error{
		"mysql": cs.StopMysqlDatabaseMigration,
		"pg":    cs.StopPgDatabaseMigration,
		"redis": cs.StopRedisDatabaseMigration,
	}

	if _, ok := stopMigrationFuncs[dbType]; !ok {
		err = fmt.Errorf("migrations not supported for database type %q", dbType)
	}

	_, err = cs.GetDatabaseMigrationStatus(ctx, c.Zone, c.Name)
	if err != nil {
		if errors.Is(err, exoapi.ErrNotFound) {
			return fmt.Errorf("migration for database %q not running in zone %q", c.Name, c.Zone)
		}
		return fmt.Errorf("failed to retrieve migration status: %s", err)
	}

	decorateAsyncOperation("Stopping Database Migration...", func() {
		err = stopMigrationFuncs[dbType](ctx, c.Zone, c.Name)
	})

	if err != nil {
		if errors.Is(err, exoapi.ErrNotFound) {
			return fmt.Errorf("migration not running in zone %q", c.Zone)
		}
		return fmt.Errorf("failed to stop migration: %s", err)
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(dbaasMigrationCmd, &dbaasMigrationStopCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
