package cmd

import (
	"fmt"
	"os"
	"strings"

	exoapi "github.com/exoscale/egoscale/v2/api"

	"github.com/exoscale/cli/table"
	"github.com/spf13/cobra"
)

type dnsListItemOutput struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
}

type dnsListOutput []dnsListItemOutput

func (o *dnsListOutput) toJSON() { outputJSON(o) }

func (o *dnsListOutput) toText() { outputText(o) }

func (o *dnsListOutput) toTable() {
	t := table.NewTable(os.Stdout)
	t.SetHeader([]string{"ID", "Name"})

	for _, i := range *o {
		t.Append([]string{
			i.ID,
			i.Name,
		})
	}

	t.Render()
}

func init() {
	dnsCmd.AddCommand(&cobra.Command{
		Use:   "list [FILTER]...",
		Short: "List domains",
		Long: fmt.Sprintf(`This command lists existing DNS Domains.
Optional patterns can be provided to filter results by ID, or name.

Supported output template annotations: %s`,
			strings.Join(outputterTemplateAnnotations(&dnsListOutput{}), ", ")),
		Aliases: gListAlias,
		RunE: func(cmd *cobra.Command, args []string) error {
			return output(listDomains(args))
		},
	})
}

func listDomains(filters []string) (outputter, error) {
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, gCurrentAccount.DefaultZone))
	domains, err := cs.ListDNSDomains(ctx, gCurrentAccount.DefaultZone)
	if err != nil {
		return nil, err
	}

	out := dnsListOutput{}

	for _, d := range domains {
		o := dnsListItemOutput{
			ID:   StrPtrFormatOutput(d.ID),
			Name: StrPtrFormatOutput(d.UnicodeName),
		}

		if len(filters) == 0 {
			out = append(out, o)
			continue
		}

		s := strings.ToLower(fmt.Sprintf("%s#%s", o.ID, o.Name))

		for _, filter := range filters {
			substr := strings.ToLower(filter)
			if strings.Contains(s, substr) {
				out = append(out, o)
				break
			}
		}
	}

	return &out, nil
}
