package golang

import (
	"context"
	"fmt"
	"log/slog"

	itbasisBuilderExec "github.com/itbasis/go-tools-builder/internal/exec"
	"github.com/itbasis/go-tools-builder/internal/installer/model"
	itbasisCoreExec "github.com/itbasis/go-tools-core/exec"
	itbasisCoreLog "github.com/itbasis/go-tools-core/log"
)

const ProviderGoKey model.ProviderKey = "go"

type GoInstaller interface {
	Install(ctx context.Context, list model.GoDependencyList) error
}

var GoInstallEmpty = (GoInstaller)(nil)

type installer struct {
	exec *itbasisCoreExec.Executable
}

func NewGoInstaller(ctx context.Context, out itbasisCoreExec.CobraOut) (GoInstaller, error) {
	var exec, err = itbasisBuilderExec.NewGoGetWithCobra(ctx, out)
	if err != nil {
		return GoInstallEmpty, err //nolint:wrapcheck // TODO
	}

	return &installer{
		exec: exec,
	}, nil
}

func (r *installer) Install(ctx context.Context, list model.GoDependencyList) error {
	for name, dependency := range list {
		slog.Info(fmt.Sprintf("install dependency: %s[%s]", name, dependency.Version))

		if err := r.exec.Execute(
			ctx,
			itbasisCoreExec.WithRerun(),
			itbasisCoreExec.WithRestoreArgsIncludePrevious(
				itbasisCoreExec.IncludePrevArgsBefore,
				"-tool",
				dependency.String(),
			),
		); err != nil {
			slog.Error("fail install dependency: "+name, itbasisCoreLog.SlogAttrError(err))
		}
	}

	return nil
}
