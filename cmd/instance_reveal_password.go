package cmd

import (
	"errors"
	"fmt"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type instanceRevealCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"reveal-password"`

	Instance string `cli-arg:"#" cli-usage:"NAME|ID"`
	Zone     string `cli-short:"z" cli-usage:"instance zone"`
}

type instanceRevealOutput struct {
	ID       string `json:"id"`
	Password string `json:"password"`
}

func (o *instanceRevealOutput) Type() string { return "Compute instance" }
func (o *instanceRevealOutput) toJSON()      { outputJSON(o) }
func (o *instanceRevealOutput) toText()      { outputText(o) }
func (o *instanceRevealOutput) toTable()     { outputTable(o) }

func (c *instanceRevealCmd) cmdAliases() []string { return nil }

func (c *instanceRevealCmd) cmdShort() string { return "Reveal the password of a Compute instance" }

func (c *instanceRevealCmd) cmdLong() string { return "" }

func (c *instanceRevealCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceRevealCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	instance, err := cs.FindInstance(ctx, c.Zone, c.Instance)
	if err != nil {
		if errors.Is(err, exoapi.ErrNotFound) {
			return fmt.Errorf("resource not found in zone %q", c.Zone)
		}
		return err
	}

	pwd, err := cs.RevealInstancePassword(ctx, c.Zone, instance)
	if err != nil {
		return err
	}

	out := instanceRevealOutput{
		ID:       *instance.ID,
		Password: pwd,
	}
	return c.outputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(instanceCmd, &instanceRevealCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
