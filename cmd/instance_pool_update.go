package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type instancePoolUpdateCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"update"`

	InstancePool string `cli-arg:"#" cli-usage:"NAME|ID"`

	AntiAffinityGroups []string          `cli-flag:"anti-affinity-group" cli-short:"a" cli-usage:"managed Compute instances Anti-Affinity Group NAME|ID (can be specified multiple times)"`
	CloudInitFile      string            `cli-flag:"cloud-init" cli-short:"c" cli-usage:"cloud-init user data configuration file path"`
	CloudInitCompress  bool              `cli-flag:"cloud-init-compress" cli-usage:"compress instance cloud-init user data"`
	DeployTarget       string            `cli-usage:"managed Compute instances Deploy Target NAME|ID"`
	Description        string            `cli-usage:"Instance Pool description"`
	Disk               int64             `cli-flag:"disk" cli-short:"d" cli-usage:"[DEPRECATED] use --disk-size"`
	DiskSize           int64             `cli-usage:"managed Compute instances disk size"`
	ElasticIPs         []string          `cli-flag:"elastic-ip" cli-short:"e" cli-usage:"managed Compute instances Elastic IP ADDRESS|ID (can be specified multiple times)"`
	IPv6               bool              `cli-flag:"ipv6" cli-short:"6" cli-usage:"enable IPv6 on managed Compute instances"`
	InstancePrefix     string            `cli-usage:"string to prefix managed Compute instances names with"`
	InstanceType       string            `cli-usage:"managed Compute instances type (format: [FAMILY.]SIZE)"`
	Keypair            string            `cli-short:"k" cli-usage:"[DEPRECATED] use --ssh-key"`
	Labels             map[string]string `cli-flag:"label" cli-usage:"Instance Pool label (format: key=value)"`
	Name               string            `cli-short:"n" cli-usage:"Instance Pool name"`
	PrivateNetworks    []string          `cli-flag:"private-network" cli-usage:"managed Compute instances Private Network NAME|ID (can be specified multiple times)"`
	Privnet            []string          `cli-short:"p" cli-usage:"[DEPRECATED] use --private-network"`
	SSHKey             string            `cli-flag:"ssh-key" cli-usage:"SSH key to deploy on managed Compute instances"`
	SecurityGroups     []string          `cli-flag:"security-group" cli-short:"s" cli-usage:"managed Compute instances Security Group NAME|ID (can be specified multiple times)"`
	ServiceOffering    string            `cli-short:"o" cli-usage:"[DEPRECATED] use --instance-type"`
	Size               int64             `cli-usage:"[DEPRECATED] use the 'exo compute instance-pool scale' command"`
	Template           string            `cli-short:"t" cli-usage:"managed Compute instances template NAME|ID"`
	TemplateFilter     string            `cli-usage:"[DEPRECATED] use --template-visibility"`
	TemplateVisibility string            `cli-usage:"instance template visibility (public|private)"`
	Zone               string            `cli-short:"z" cli-usage:"Instance Pool zone"`
}

func (c *instancePoolUpdateCmd) cmdAliases() []string { return nil }

func (c *instancePoolUpdateCmd) cmdShort() string { return "Update an Instance Pool" }

func (c *instancePoolUpdateCmd) cmdLong() string {
	return fmt.Sprintf(`This command updates an Instance Pool.

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&instancePoolShowOutput{}), ", "),
	)
}

func (c *instancePoolUpdateCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	// TODO: remove this once the `--disk` flag is retired.
	if cmd.Flags().Changed("disk") {
		cmd.PrintErr(`**********************************************************************
WARNING: flag "--disk" has been deprecated and will be removed in a
future release, please use "--disk-size" instead.
**********************************************************************
`)
		if !cmd.Flags().Changed("disk-size") {
			diskFlag := cmd.Flags().Lookup("disk")
			if err := cmd.Flags().Set("disk-size", fmt.Sprint(diskFlag.Value.String())); err == nil {
				return err
			}
		}
	}

	// TODO: remove this once the `--keypair` flag is retired.
	if cmd.Flags().Changed("keypair") {
		cmd.PrintErr(`**********************************************************************
WARNING: flag "--keypair" has been deprecated and will be removed in
a future release, please use "--ssh-key" instead.
**********************************************************************
`)
		if !cmd.Flags().Changed("ssh-key") {
			keypairFlag := cmd.Flags().Lookup("keypair")
			if err := cmd.Flags().Set("ssh-key", keypairFlag.Value.String()); err != nil {
				return err
			}
		}
	}

	// TODO: remove this once the `--privnet` flag is retired.
	if cmd.Flags().Changed("privnet") {
		cmd.PrintErr(`**********************************************************************
WARNING: flag "--privnet" has been deprecated and will be removed in
a future release, please use "--private-network" instead.
**********************************************************************
`)
		if !cmd.Flags().Changed("private-network") {
			privnetFlag := cmd.Flags().Lookup("privnet")
			if err := cmd.Flags().Set(
				"private-network",
				strings.Trim(privnetFlag.Value.String(), "[]"),
			); err != nil {
				return err
			}
		}
	}

	// TODO: remove this once the `--service-offering` flag is retired.
	if cmd.Flags().Changed("service-offering") {
		cmd.PrintErr(`**********************************************************************
WARNING: flag "--service-offering" has been deprecated and will be removed
in a future release, please use "--instance-type" instead.
**********************************************************************
`)
		if !cmd.Flags().Changed("instance-type") {
			serviceOfferingFlag := cmd.Flags().Lookup("service-offering")
			if err := cmd.Flags().Set("instance-type", serviceOfferingFlag.Value.String()); err != nil {
				return err
			}
		}
	}

	// TODO: remove this once the `--template-filter` flag is retired.
	if cmd.Flags().Changed("template-filter") {
		cmd.PrintErr(`**********************************************************************
WARNING: flag "--template-filter" has been deprecated and will be removed in
a future release, please use "--template-visibility" instead.
**********************************************************************
`)
		if !cmd.Flags().Changed("template-visibility") {
			templateFilterFlag := cmd.Flags().Lookup("template-filter")
			if err := cmd.Flags().Set("template-visibility", templateFilterFlag.Value.String()); err != nil {
				return err
			}
		}
	}

	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *instancePoolUpdateCmd) cmdRun(cmd *cobra.Command, _ []string) error {
	var updated bool

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	instancePool, err := cs.FindInstancePool(ctx, c.Zone, c.InstancePool)
	if err != nil {
		if errors.Is(err, exoapi.ErrNotFound) {
			return fmt.Errorf("resource not found in zone %q", c.Zone)
		}
		return err
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.AntiAffinityGroups)) {
		antiAffinityGroupIDs := make([]string, len(c.AntiAffinityGroups))
		for i, v := range c.AntiAffinityGroups {
			antiAffinityGroup, err := cs.FindAntiAffinityGroup(ctx, c.Zone, v)
			if err != nil {
				return fmt.Errorf("error retrieving Anti-Affinity Group: %w", err)
			}
			antiAffinityGroupIDs[i] = *antiAffinityGroup.ID
		}
		instancePool.AntiAffinityGroupIDs = &antiAffinityGroupIDs
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.DeployTarget)) {
		deployTarget, err := cs.FindDeployTarget(ctx, c.Zone, c.DeployTarget)
		if err != nil {
			return fmt.Errorf("error retrieving Deploy Target: %w", err)
		}
		instancePool.DeployTargetID = deployTarget.ID
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.Description)) {
		instancePool.Description = &c.Description
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.DiskSize)) {
		instancePool.DiskSize = &c.DiskSize
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.ElasticIPs)) {
		elasticIPIDs := make([]string, len(c.ElasticIPs))
		for i, v := range c.ElasticIPs {
			elasticIP, err := cs.FindElasticIP(ctx, c.Zone, v)
			if err != nil {
				return fmt.Errorf("error retrieving Elastic IP: %w", err)
			}
			elasticIPIDs[i] = *elasticIP.ID
		}
		instancePool.ElasticIPIDs = &elasticIPIDs
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.InstancePrefix)) {
		instancePool.InstancePrefix = &c.InstancePrefix
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.IPv6)) {
		instancePool.IPv6Enabled = &c.IPv6
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.Labels)) {
		instancePool.Labels = &c.Labels
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.Name)) {
		instancePool.Name = &c.Name
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.PrivateNetworks)) {
		privateNetworkIDs := make([]string, len(c.PrivateNetworks))
		for i, v := range c.PrivateNetworks {
			privateNetwork, err := cs.FindPrivateNetwork(ctx, c.Zone, v)
			if err != nil {
				return fmt.Errorf("error retrieving Private Network: %w", err)
			}
			privateNetworkIDs[i] = *privateNetwork.ID
		}
		instancePool.PrivateNetworkIDs = &privateNetworkIDs
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.SecurityGroups)) {
		securityGroupIDs := make([]string, len(c.SecurityGroups))
		for i, v := range c.SecurityGroups {
			securityGroup, err := cs.FindSecurityGroup(ctx, c.Zone, v)
			if err != nil {
				return fmt.Errorf("error retrieving Security Group: %w", err)
			}
			securityGroupIDs[i] = *securityGroup.ID
		}
		instancePool.SecurityGroupIDs = &securityGroupIDs
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.InstanceType)) {
		instanceType, err := cs.FindInstanceType(ctx, c.Zone, c.InstanceType)
		if err != nil {
			return fmt.Errorf("error retrieving instance type: %w", err)
		}
		instancePool.InstanceTypeID = instanceType.ID
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.SSHKey)) {
		instancePool.SSHKey = &c.SSHKey
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.Template)) {
		template, err := cs.FindTemplate(ctx, c.Zone, c.Template, c.TemplateVisibility)
		if err != nil {
			return fmt.Errorf(
				"no template %q found with visibility %s in zone %s",
				c.Template,
				c.TemplateVisibility,
				c.Zone,
			)
		}
		instancePool.TemplateID = template.ID
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.CloudInitFile)) {
		userData, err := getUserDataFromFile(c.CloudInitFile, c.CloudInitCompress)
		if err != nil {
			return fmt.Errorf("error parsing cloud-init user data: %w", err)
		}
		instancePool.UserData = &userData
		updated = true
	}

	if updated {
		decorateAsyncOperation(fmt.Sprintf("Updating Instance Pool %q...", c.InstancePool), func() {
			if err = cs.UpdateInstancePool(ctx, c.Zone, instancePool); err != nil {
				return
			}
		})
		if err != nil {
			return err
		}
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.Size)) {
		_, _ = fmt.Fprintln(
			os.Stderr,
			`WARNING: the "--size" flag is deprecated and replaced by the `+
				`"exo compute instance-pool scale" command, it will be removed `+
				`in a future version.`,
		)

		decorateAsyncOperation(fmt.Sprintf("Scaling Instance Pool %q...", c.InstancePool), func() {
			err = cs.ScaleInstancePool(ctx, c.Zone, instancePool, c.Size)
		})
	}

	if !gQuiet {
		return (&instancePoolShowCmd{
			cliCommandSettings: c.cliCommandSettings,
			Zone:               c.Zone,
			InstancePool:       *instancePool.ID,
		}).cmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(instancePoolCmd, &instancePoolUpdateCmd{
		cliCommandSettings: defaultCLICmdSettings(),

		TemplateVisibility: defaultTemplateVisibility,
	}))

	// FIXME: remove this someday.
	cobra.CheckErr(registerCLICommand(deprecatedInstancePoolCmd, &instancePoolUpdateCmd{
		cliCommandSettings: defaultCLICmdSettings(),

		TemplateVisibility: defaultTemplateVisibility,
	}))
}
