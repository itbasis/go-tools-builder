package installer

import (
	"context"
	"log/slog"

	"github.com/itbasis/go-tools-builder/internal/installer/model"
	"github.com/itbasis/go-tools-builder/internal/installer/providers/github"
	"github.com/itbasis/go-tools-builder/internal/installer/providers/golang"
	itbasisCoreExec "github.com/itbasis/go-tools-core/exec"
	itbasisCoreOption "github.com/itbasis/go-tools-core/option"
)

type Installer struct {
	dependencies model.Dependencies

	cobraOut itbasisCoreExec.CobraOut

	providers map[model.ProviderKey]any
}

func NewInstaller(ctx context.Context, cobraOut itbasisCoreExec.CobraOut, opts ...Option) (*Installer, error) {
	installer := &Installer{
		cobraOut:  cobraOut,
		providers: map[model.ProviderKey]any{},
	}

	if err := itbasisCoreOption.ApplyOptions(ctx, installer, opts, nil); err != nil {
		return nil, err //nolint:wrapcheck // TODO
	}

	return installer, nil
}

func (r *Installer) Install(ctx context.Context) error {
	if len(r.dependencies.Go) > 0 {
		if err := r.installGo(ctx); err != nil {
			return err
		}
	}

	if len(r.dependencies.Github) > 0 {
		if err := r.installGitHub(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (r *Installer) installGo(ctx context.Context) error {
	var (
		err       error
		installer golang.GoInstaller
	)

	if value, exist := r.providers[golang.ProviderGoKey]; !exist {
		installer, err = golang.NewGoInstaller(ctx, r.cobraOut)
		if err != nil {
			return err //nolint:wrapcheck // TODO
		}

		r.providers[golang.ProviderGoKey] = installer
	} else {
		installer = value.(golang.GoInstaller)
	}

	slog.Info("installing dependencies with provider: " + string(golang.ProviderGoKey))

	if err = installer.Install(ctx, r.dependencies.Go); err != nil {
		return err //nolint:wrapcheck // TODO
	}

	return nil
}

func (r *Installer) installGitHub(ctx context.Context) error {
	var (
		err       error
		installer github.Installer
	)

	if value, exist := r.providers[github.ProviderGithubKey]; !exist {
		installer, err = github.NewGithubInstaller()
		if err != nil {
			return err //nolint:wrapcheck // TODO
		}

		r.providers[github.ProviderGithubKey] = installer
	} else {
		installer = value.(github.Installer)
	}

	slog.Info("installing dependencies with provider: " + string(github.ProviderGithubKey))

	if err = installer.Install(ctx, r.dependencies.Github); err != nil {
		return err //nolint:wrapcheck // TODO
	}

	return nil
}
