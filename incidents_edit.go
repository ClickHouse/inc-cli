package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/alexeldeib/incli/client"
	kitlog "github.com/go-kit/log"
	"github.com/spf13/cobra"
)

func NewPatchIncidentsCommand() *cobra.Command {
	opts := &PatchIncidentOptions{}

	cmd := &cobra.Command{
		Use:   "edit",
		Short: "edit an incident",
		Run: func(cmd *cobra.Command, args []string) {
			ctx, logger, cl, err := setup()
			if err != nil {
				fmt.Printf("failed to setup: %s", err)
				os.Exit(1)
			}

			if err := opts.Run(ctx, logger, cl); err != nil {
				logger.Log("msg", "failed to run", "error", err)
				os.Exit(1)
			}
		},
	}

	cmd.Flags().StringVar(&opts.incidentID, "id", "", "incident ID, e.g. 01HE6...")
	cmd.Flags().IntVar(&opts.incidentReference, "ref", -1, "incident reference number, e.g. 27 for INC-27")
	cmd.Flags().StringSliceVar(&opts.customFields, "field", nil, "custom field to patch, e.g. --field foo=bar --field baz=qux. --field foo=bar=baz sets field `foo` to `bar=baz`")

	return cmd
}

type PatchIncidentOptions struct {
	incidentReference int
	incidentID        string
	customFields      []string
}

func (o *PatchIncidentOptions) Run(ctx context.Context, logger kitlog.Logger, cl *client.ClientWithResponses) error {
	if o.incidentReference != -1 && o.incidentID != "" {
		return fmt.Errorf("exactly one of --id or --ref may be specified")
	}

	if o.incidentReference == -1 && o.incidentID == "" {
		return fmt.Errorf("exactly one of --id or --ref must be specified")
	}

	if o.incidentReference != -1 && o.incidentReference < 0 {
		return fmt.Errorf("incident --ref must be positive integer: %q", o.incidentReference)
	}

	if len(o.customFields) == 0 {
		return fmt.Errorf("at least one edit field must be specified")
	}

	customFieldsMap := map[string]string{}
	for _, v := range o.customFields {
		parts := strings.Split(v, "=")
		if len(parts) < 1 {
			return fmt.Errorf("invalid custom field: %q", v)
		}

		switch len(parts) {
		case 1:
			customFieldsMap[parts[0]] = "" // field reset
		case 2:
			customFieldsMap[parts[0]] = parts[1]
		default: // field value contains equal signs
			customFieldsMap[parts[0]] = strings.Join(parts[1:], "=")
		}
	}

	if o.incidentReference > 0 {
		res, err := EditIncidentByReferenceNumber(ctx, logger, cl, o.incidentReference, customFieldsMap)
		if err != nil {
			return fmt.Errorf("failed to edit incident: %q", err)
		}

		if err := serialize(res); err != nil {
			return fmt.Errorf("failed to marshal json: %q", err)
		}

		return nil
	}

	incident, err := ShowIncidentByID(ctx, logger, cl, o.incidentID)
	if err != nil {
		return fmt.Errorf("failed to list incidents: %q", err)
	}

	if err := serialize(incident); err != nil {
		return fmt.Errorf("failed to marshal json: %q", err)
	}

	return nil
}
