package commands

import (
	"errors"
	"fmt"
	"os"

	"github.com/bearer/curio/pkg/commands/artifact"
	"github.com/bearer/curio/pkg/flag"
	"github.com/bearer/curio/pkg/util/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/xerrors"
)

const (
	scanTemplate = `Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}
Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}
Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}
Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}

{{if .HasAvailableLocalFlags}}
%s
{{end}}
`
)

var scanFlags = &flag.Flags{
	ScanFlagGroup:    flag.NewScanFlagGroup(),
	WorkerFlagGroup:  flag.NewWorkerFlagGroup(),
	ReportFlagGroup:  flag.NewReportFlagGroup(),
	GeneralFlagGroup: flag.NewGeneralFlagGroup(),
}

func NewScanCommand() *cobra.Command {

	cmd := &cobra.Command{
		Use:     "scan [flags] PATH",
		Aliases: []string{"s"},
		Short:   "Scan git repository",
		Example: `  # Scan a local project including language-specific files
  $ curio s /path/to/your_project
  # Scan a single file
  $ curio s ./curio-ci-test/Pipfile.lock`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := scanFlags.Bind(cmd); err != nil {
				return xerrors.Errorf("flag bind error: %w", err)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := scanFlags.Bind(cmd); err != nil {
				return xerrors.Errorf("flag bind error: %w", err)
			}

			configPath := viper.GetString(flag.ConfigFileFlag.ConfigName)

			if err := readConfig(configPath); err != nil {
				return err
			}

			options, err := scanFlags.ToOptions(args, outputWriter)
			if err != nil {
				return xerrors.Errorf("flag error: %w", err)
			}

			output.Setup(options)

			if options.Target == "" {
				return cmd.Help()
			}

			return artifact.Run(cmd.Context(), options, artifact.TargetFilesystem)
		},
		SilenceErrors: false,
		SilenceUsage:  false,
	}

	scanFlags.AddFlags(cmd)
	cmd.SetUsageTemplate(fmt.Sprintf(scanTemplate, scanFlags.Usages(cmd)))

	return cmd
}

func readConfig(configFile string) error {
	// Read from config
	viper.SetConfigFile(configFile)
	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}

		return fmt.Errorf("config file %q loading error: %s", configFile, err)
	}

	output.StdErrLogger().Msgf("Loaded %s configuration file", configFile)

	return nil
}