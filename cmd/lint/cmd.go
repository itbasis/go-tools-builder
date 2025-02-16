package lint

import (
	builderCmd "github.com/itbasis/go-tools-builder/internal/cmd"
	itbasisCoreCmd "github.com/itbasis/go-tools-core/cmd"
	itbasisCoreExec "github.com/itbasis/go-tools-core/exec"
	itbasisCoreOs "github.com/itbasis/go-tools-core/os"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	_flagSkipEditorConfigChecker bool
	_flagSkipGolangCiLint        bool
)

func NewLintCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:    itbasisCoreCmd.BuildUse("lint", builderCmd.UseArgPackages),
		Args:   cobra.MatchAll(cobra.OnlyValidArgs, cobra.MaximumNArgs(1)),
		PreRun: itbasisCoreCmd.LogCommand,
		Run:    _run,
	}

	cmd.Flags().BoolVar(&_flagSkipEditorConfigChecker, "skip-editorconfig-checker", false, "skip editor config checker")
	cmd.Flags().BoolVar(&_flagSkipGolangCiLint, "skip-golangci-lint", false, "skip golangci-lint")

	return cmd
}

func _run(cmd *cobra.Command, args []string) {
	withCobraOut := itbasisCoreExec.WithCobraOut(cmd)

	if !_flagSkipEditorConfigChecker && itbasisCoreOs.BeARegularFile(".editorconfig") {
		itbasisCoreCmd.RequireNoError(cmd, _execEditorConfigChecker(withCobraOut))
	}

	if !_flagSkipGolangCiLint {
		itbasisCoreCmd.RequireNoError(cmd, _execGolangCiLint(builderCmd.ArgPackages(builderCmd.DefaultPackages, args), withCobraOut))
	}
}

func _execEditorConfigChecker(opts ...itbasisCoreExec.Option) error {
	executable, err := itbasisCoreExec.NewExecutable("editorconfig-checker", opts...)
	if err != nil {
		return errors.Wrap(err, itbasisCoreExec.ErrFailedExecuteCommand.Error())
	}

	if err := executable.Execute(); err != nil {
		return errors.Wrap(err, itbasisCoreExec.ErrFailedExecuteCommand.Error())
	}

	return nil
}

func _execGolangCiLint(lintPackages string, opts ...itbasisCoreExec.Option) error {
	executable, err := itbasisCoreExec.NewExecutable(
		"golangci-lint",
		append(
			[]itbasisCoreExec.Option{itbasisCoreExec.WithArgs("run", lintPackages)},
			opts...,
		)...,
	)
	if err != nil {
		return errors.Wrap(err, itbasisCoreExec.ErrFailedExecuteCommand.Error())
	}

	return errors.Wrap(executable.Execute(), itbasisCoreExec.ErrFailedExecuteCommand.Error())
}
