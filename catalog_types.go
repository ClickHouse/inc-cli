package main

import (
	"context"
	"fmt"
	"os"

	"github.com/alexeldeib/incli/client"
	kitlog "github.com/go-kit/log"
	"github.com/spf13/cobra"
)

func NewGetCatalogTypesCommand() *cobra.Command {
	opts := &GetCatalogTypesOptions{}
	cmd := &cobra.Command{
		Use:   "get",
		Short: "get one or all catalog types, by name or id",
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

	cmd.Flags().StringVar(&opts.catalogTypeName, "name", "", "catalog type name, e.g. PagerdutyService")
	cmd.Flags().StringVar(&opts.catalogTypeID, "id", "", "catalog type id, e.g. 01HE6...")

	return cmd
}

type GetCatalogTypesOptions struct {
	catalogTypeName string
	catalogTypeID   string
}

func (o *GetCatalogTypesOptions) Run(ctx context.Context, logger kitlog.Logger, cl *client.ClientWithResponses) error {
	if o.catalogTypeName != "" && o.catalogTypeID != "" {
		return fmt.Errorf("exactly one of --type-name or --type-id may be specified")
	}

	if o.catalogTypeID != "" {
		res, err := FindCatalogTypeByID(ctx, logger, cl, o.catalogTypeID)
		if err != nil {
			return fmt.Errorf("failed to find catalog entry: %s", err)
		}

		if err := serialize(res); err != nil {
			return fmt.Errorf("failed to marshal json: %q", err)
		}

		return nil
	} else if o.catalogTypeName != "" {
		res, err := FindCatalogTypeByName(ctx, logger, cl, o.catalogTypeName)
		if err != nil {
			return fmt.Errorf("failed to find catalog entry: %s", err)
		}

		if err := serialize(res); err != nil {
			return fmt.Errorf("failed to marshal json: %q", err)
		}

		return nil
	}

	res, err := ListAllCatalogTypes(ctx, logger, cl)
	if err != nil {
		return fmt.Errorf("failed to find catalog entry: %s", err)
	}

	if err := serialize(res); err != nil {
		return fmt.Errorf("failed to marshal json: %q", err)
	}

	return nil
}
