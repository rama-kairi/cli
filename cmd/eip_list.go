package cmd

import (
	"os"

	"github.com/exoscale/egoscale"
	"github.com/exoscale/egoscale/cmd/exo/table"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var eipListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List elastic IP",
	Aliases: gListAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		zone, err := cmd.Flags().GetString("zone")
		if err != nil {
			return err
		}
		table := table.NewTable(os.Stdout)
		table.SetHeader([]string{"zone", "IP", "ID"})
		if err := listIPs(zone, table); err != nil {
			return err
		}
		table.Render()
		return nil
	},
}

func listIPs(zone string, table *table.Table) error {
	zReq := egoscale.IPAddress{}

	if zone != "" {
		var err error
		zReq.ZoneID, err = getZoneIDByName(cs, zone)
		if err != nil {
			return err
		}
		zReq.IsElastic = true
		ips, err := cs.List(&zReq)
		if err != nil {
			return err
		}

		for _, ipaddr := range ips {
			ip := ipaddr.(*egoscale.IPAddress)
			table.Append([]string{ip.ZoneName, ip.IPAddress.String(), ip.ID})
		}
		return nil
	}

	zones := &egoscale.Zone{}
	zs, err := cs.List(zones)
	if err != nil {
		return err
	}

	for _, z := range zs {
		zID := z.(*egoscale.Zone).Name
		if err := listIPs(zID, table); err != nil {
			return err
		}
	}
	return nil
}

func init() {
	eipListCmd.Flags().StringP("zone", "z", "", "Show IPs from given zone")
	eipCmd.AddCommand(eipListCmd)
}
