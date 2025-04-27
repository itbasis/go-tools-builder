package dependencies

import (
	_ "embed"
	"log/slog"

	itbasisBuilderExec "github.com/itbasis/go-tools-builder/internal/exec"
	builderInstaller "github.com/itbasis/go-tools-builder/internal/installer"
	itbasisCoreCmd "github.com/itbasis/go-tools-core/cmd"
	itbasisCoreExec "github.com/itbasis/go-tools-core/exec"
	"github.com/spf13/cobra"
)

//go:embed dependencies.json
var _defaultDependencies []byte

var (
	_flagDependenciesFile string
	_flagShow             bool
)

func NewDependenciesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:    itbasisCoreCmd.BuildUse("dependencies"),
		Short:  "Install dependencies",
		Args:   cobra.NoArgs,
		PreRun: itbasisCoreCmd.LogCommand,
		RunE:   _runE,
	}

	flags := cmd.Flags()

	flags.StringVarP(
		&_flagDependenciesFile,
		"dependencies-file",
		"f",
		"",
		"dependencies file path. If not specified, the embedded list will be used",
	)
	flags.BoolVar(&_flagShow, "show-default", false, "show default dependencies for install")

	return cmd
}

func _runE(cmd *cobra.Command, _ []string) error {
	if _flagShow {
		_, err := cmd.OutOrStdout().Write(_defaultDependencies)

		return err //nolint:wrapcheck // TODO
	}

	var optionDependencies builderInstaller.Option

	if _flagDependenciesFile != "" {
		slog.Info("using dependencies file: " + _flagDependenciesFile)

		optionDependencies = builderInstaller.WithFile(_flagDependenciesFile)
	} else {
		optionDependencies = builderInstaller.WithJSONData(_defaultDependencies)
	}

	var (
		ctx                     = cmd.Context()
		installer, errInstaller = builderInstaller.NewInstaller(ctx, cmd, optionDependencies)
		goModTidy, errGoMod     = itbasisBuilderExec.NewGoModWithCobra(ctx, cmd)
	)

	itbasisCoreCmd.RequireNoError(cmd, errInstaller)
	itbasisCoreCmd.RequireNoError(cmd, errGoMod)

	itbasisCoreCmd.RequireNoError(cmd, installer.Install(ctx))
	itbasisCoreCmd.RequireNoError(cmd, goModTidy.Execute(ctx,
		itbasisCoreExec.WithRestoreArgsIncludePrevious(itbasisCoreExec.IncludePrevArgsBefore, "tidy"),
	))

	return nil
}
