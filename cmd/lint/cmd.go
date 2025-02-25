package lint

import (
	"context"
	"os"

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
	var (
		ctx = cmd.Context()
		pwd = itbasisCoreOs.Pwd()

		withWorkDir  = itbasisCoreExec.WithWorkDir(pwd)
		withCobraOut = itbasisCoreExec.WithCobraOut(cmd)
	)

	if !_flagSkipEditorConfigChecker && itbasisCoreOs.BeARegularFile(os.DirFS(pwd), ".editorconfig") {
		itbasisCoreCmd.RequireNoError(
			cmd,
			_execEditorConfigChecker(ctx, withCobraOut, withWorkDir),
		)
	}

	if !_flagSkipGolangCiLint {
		itbasisCoreCmd.RequireNoError(
			cmd,
			_execGolangCiLint(ctx, builderCmd.ArgPackages(builderCmd.DefaultPackages, args), withCobraOut, withWorkDir),
		)
	}
}

func _execEditorConfigChecker(ctx context.Context, opts ...itbasisCoreExec.Option) error {
	executable, err := itbasisCoreExec.NewExecutable(ctx, "editorconfig-checker", opts...)
	if err != nil {
		return errors.Wrap(err, itbasisCoreExec.ErrFailedExecuteCommand.Error())
	}

	if err := executable.Execute(ctx); err != nil {
		return errors.Wrap(err, itbasisCoreExec.ErrFailedExecuteCommand.Error())
	}

	return nil
}

func _execGolangCiLint(ctx context.Context, lintPackages string, opts ...itbasisCoreExec.Option) error {
	executable, err := itbasisCoreExec.NewExecutable(
		ctx,
		"golangci-lint",
		append(
			[]itbasisCoreExec.Option{itbasisCoreExec.WithArgs("run", lintPackages)},
			opts...,
		)...,
	)
	if err != nil {
		return errors.Wrap(err, itbasisCoreExec.ErrFailedExecuteCommand.Error())
	}

	return errors.Wrap(executable.Execute(ctx), itbasisCoreExec.ErrFailedExecuteCommand.Error())
}
