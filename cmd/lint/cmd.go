package lint

import (
	"context"
	"log/slog"
	"os"

	builderCmd "github.com/itbasis/go-tools-builder/internal/cmd"
	itbasisBuilderExec "github.com/itbasis/go-tools-builder/internal/exec"
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
			_execEditorConfigChecker(cmd),
		)
	}

	if !_flagSkipGolangCiLint {
		itbasisCoreCmd.RequireNoError(
			cmd,
			_execGolangCiLint(ctx, builderCmd.ArgPackages(builderCmd.DefaultPackages, args), withCobraOut, withWorkDir),
		)
	}
}

func _execEditorConfigChecker(cmd *cobra.Command) error {
	const tool = "editorconfig-checker"

	slog.Info("running tool: " + tool)

	var (
		ctx             = cmd.Context()
		exec, errGoTool = itbasisBuilderExec.NewGoToolWithCobra(ctx, cmd)
	)

	itbasisCoreCmd.RequireNoError(cmd, errGoTool)

	if err := exec.Execute(ctx,
		itbasisCoreExec.WithRerun(),
		itbasisCoreExec.WithRestoreArgsIncludePrevious(itbasisCoreExec.IncludePrevArgsBefore, tool),
	); err != nil {
		return errors.Wrap(err, itbasisCoreExec.ErrFailedExecuteCommand.Error())
	}

	return nil
}

func _execGolangCiLint(ctx context.Context, lintPackages string, opts ...itbasisCoreExec.Option) error {
	const tool = "golangci-lint"

	slog.Info("running tool: " + tool)

	executable, err := itbasisCoreExec.NewExecutable(
		ctx,
		tool,
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
