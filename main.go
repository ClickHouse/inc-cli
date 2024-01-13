package main

import (
	"context"
	"fmt"
	stdlog "log"
	"os"
	"os/signal"
	"syscall"

	kitlog "github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/alexeldeib/incli/client"
)

func main() {
	if err := NewRootCommand().Execute(); err != nil {
		fmt.Printf("%+#v\n", err)
		os.Exit(1)
	}
}

func setup() (context.Context, kitlog.Logger, *client.ClientWithResponses, error) {
	ctx, cancel := context.WithCancel(context.Background())

	logger := kitlog.NewLogfmtLogger(kitlog.NewSyncWriter(os.Stderr))
	logger = level.NewFilter(logger, level.AllowInfo())
	logger = kitlog.With(logger, "timestamp", kitlog.DefaultTimestampUTC, "caller", kitlog.DefaultCaller)
	stdlog.SetOutput(kitlog.NewStdlibAdapter(logger))

	// Setup signal handling.
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	go func() {
		<-sigc
		cancel()
		<-sigc
		logger.Log("msg", "received second signal, exiting immediately")
		os.Exit(1)
	}()

	apiKey := os.Getenv("INC_API_KEY")

	if apiKey == "" {
		return nil, nil, nil, fmt.Errorf("INC_API_KEY must be set")
	}

	cl, err := client.New(ctx, apiKey, "https://api.incident.io", "ace-cohere-cli")
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "creating client")
	}

	return ctx, logger, cl, nil
}

func NewRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use: "inc",
	}
	root.AddCommand()
	root.AddCommand(NewIncidentsCommand())
	root.AddCommand(NewCatalogCommand())

	return root
}

func NewCatalogCommand() *cobra.Command {
	root := &cobra.Command{
		Use: "catalog",
	}
	entries := &cobra.Command{
		Use: "entries",
	}

	types := &cobra.Command{
		Use: "types",
	}

	root.AddCommand()
	root.AddCommand(entries)
	root.AddCommand(types)
	entries.AddCommand(NewGetCatalogEntriesCommand())
	types.AddCommand(NewGetCatalogTypesCommand())

	return root
}

func NewIncidentsCommand() *cobra.Command {
	root := &cobra.Command{
		Use:     "incidents",
		Aliases: []string{"inc"},
	}
	root.AddCommand()
	root.AddCommand(NewGetIncidentCommand())
	root.AddCommand(NewPatchIncidentsCommand())

	return root
}
