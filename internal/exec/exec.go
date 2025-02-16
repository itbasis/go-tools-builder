package exec

import (
	"context"

	itbasisCoreExec "github.com/itbasis/go-tools-core/exec"
)

func NewGoExecutable(ctx context.Context, opts ...itbasisCoreExec.Option) (*itbasisCoreExec.Executable, error) {
	return itbasisCoreExec.NewExecutable(ctx, "go", opts...) //nolint:wrapcheck // TODO
}

func NewGoInstallWithCobra(ctx context.Context, cobraOut itbasisCoreExec.CobraOut) (*itbasisCoreExec.Executable, error) {
	return NewGoExecutable(
		ctx,
		itbasisCoreExec.WithArgs("install"),
		itbasisCoreExec.WithCobraOut(cobraOut),
	)
}

func NewGoGetWithCobra(ctx context.Context, cobraOut itbasisCoreExec.CobraOut) (*itbasisCoreExec.Executable, error) {
	return NewGoExecutable(
		ctx,
		itbasisCoreExec.WithArgs("get"),
		itbasisCoreExec.WithCobraOut(cobraOut),
	)
}

func NewGoModWithCobra(ctx context.Context, cobraOut itbasisCoreExec.CobraOut) (*itbasisCoreExec.Executable, error) {
	return NewGoExecutable(
		ctx,
		itbasisCoreExec.WithArgs("mod"),
		itbasisCoreExec.WithCobraOut(cobraOut),
	)
}

func NewGoToolWithCobra(ctx context.Context, cobraOut itbasisCoreExec.CobraOut) (*itbasisCoreExec.Executable, error) {
	return NewGoExecutable(
		ctx,
		itbasisCoreExec.WithArgs("tool"),
		itbasisCoreExec.WithCobraOut(cobraOut),
	)
}

func NewGoGenerateWithCobra(ctx context.Context, cobraOut itbasisCoreExec.CobraOut) (*itbasisCoreExec.Executable, error) {
	return NewGoExecutable(
		ctx,
		itbasisCoreExec.WithArgs("generate"),
		itbasisCoreExec.WithCobraOut(cobraOut),
	)
}

func NewGoBuildWithCobra(ctx context.Context, cobraOut itbasisCoreExec.CobraOut) (*itbasisCoreExec.Executable, error) {
	return NewGoExecutable(
		ctx,
		itbasisCoreExec.WithArgs("build"),
		itbasisCoreExec.WithCobraOut(cobraOut),
	)
}

func NewGoRunWithCobra(ctx context.Context, cobraOut itbasisCoreExec.CobraOut) (*itbasisCoreExec.Executable, error) {
	return NewGoExecutable(
		ctx,
		itbasisCoreExec.WithArgs("run"),
		itbasisCoreExec.WithCobraOut(cobraOut),
	)
}
