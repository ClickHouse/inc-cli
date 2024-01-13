package main

import (
	"context"
	"fmt"
	"os"

	"github.com/alexeldeib/incli/client"
	kitlog "github.com/go-kit/log"
	"github.com/spf13/cobra"
)

type GetIncidentOptions struct {
	incidentReference int
	incidentID        string
}

func (o *GetIncidentOptions) Run(ctx context.Context, logger kitlog.Logger, cl *client.ClientWithResponses) error {
	if o.incidentReference != -1 && o.incidentID != "" {
		return fmt.Errorf("only one of --id or --ref may be specified")
	}

	if o.incidentReference != -1 && o.incidentReference < 0 {
		return fmt.Errorf("incident --ref must be positive integer: %q", o.incidentReference)
	}

	if o.incidentReference > 0 {
		incident, err := ShowIncidentByReference(ctx, logger, cl, o.incidentReference)
		if err != nil {
			return fmt.Errorf("failed to list incidents: %s", err)
		}

		if err := serialize(incident); err != nil {
			return fmt.Errorf("failed to marshal json: %s", err)
		}

		return nil
	}

	if o.incidentID != "" {
		incident, err := ShowIncidentByID(ctx, logger, cl, o.incidentID)
		if err != nil {
			return fmt.Errorf("failed to list incidents: %s", err)
		}

		if err := serialize(incident); err != nil {
			return fmt.Errorf("failed to marshal json: %s", err)
		}

		return nil
	}

	incidents, err := ListAllIncidents(ctx, logger, cl)
	if err != nil {
		return fmt.Errorf("failed to list incidents: %s", err)
	}

	if err := serialize(incidents); err != nil {
		return fmt.Errorf("failed to marshal json: %s", err)
	}

	return nil
}

func NewGetIncidentCommand() *cobra.Command {
	opts := &GetIncidentOptions{}
	cmd := &cobra.Command{
		Use:   "get",
		Short: "get one or all incidents",
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

	return cmd
}
