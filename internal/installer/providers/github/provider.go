package github

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"syscall"

	"github.com/google/go-github/v68/github"
	itbasisBuilderHttp "github.com/itbasis/go-tools-builder/internal/http"
	"github.com/itbasis/go-tools-builder/internal/installer/model"
	"github.com/pkg/errors"
	"golift.io/xtractr"
)

const ProviderGithubKey model.ProviderKey = "github"

type Installer interface {
	Install(ctx context.Context, list model.GithubDependencyList) error
}

var InstallerEmpty = (Installer)(nil)

var excludeFiles = []string{"LICENSE", "README.md"}

type installer struct {
	githubClient *github.Client
	httpClient   *http.Client

	binPath string
}

func NewGithubInstaller() (Installer, error) {
	var binPath string

	if s := os.Getenv("GOBIN"); s != "" {
		binPath = s
	} else {
		binPath = filepath.Join(os.Getenv("GOPATH"), "bin")
	}

	return &installer{
		githubClient: github.NewClient(nil),
		httpClient:   itbasisBuilderHttp.NewHTTPClient(),
		binPath:      binPath,
	}, nil
}

func (r *installer) Install(ctx context.Context, list model.GithubDependencyList) error {
	for name, dependency := range list {
		if err := r.installDependency(ctx, name, dependency); err != nil {
			return err
		}
	}

	return nil
}

func (r *installer) installDependency(ctx context.Context, name model.DependencyName, dependency model.GithubDependency) error {
	slog.Info(fmt.Sprintf("install dependency: %s[%s]", name, dependency.Version))

	ghAsset, errGetAsset := r.getGitHubAsset(ctx, dependency)
	if errGetAsset != nil {
		return errGetAsset
	}

	var tmpFile, errCreate = os.CreateTemp("", fmt.Sprintf("tmp.%s-*%s", name, filepath.Ext(ghAsset.GetName())))
	if errCreate != nil {
		return errors.WithMessage(errCreate, "fail create temporary file")
	}

	defer func() {
		_ = tmpFile.Close()
		_ = os.RemoveAll(tmpFile.Name())
	}()

	if err := r.downloadAsset(ctx, ghAsset, tmpFile); err != nil {
		return err
	}

	return r.extractAsset(name, tmpFile)
}

func (r *installer) getGitHubAsset(ctx context.Context, dependency model.GithubDependency) (*github.ReleaseAsset, error) {
	var (
		githubRelease    *github.RepositoryRelease
		errGetRepository error
	)

	switch dependency.Version {
	case model.VersionLatest:
		githubRelease, _, errGetRepository = r.githubClient.Repositories.GetLatestRelease(ctx, dependency.Owner, dependency.Repo)

	default:
		githubRelease, _, errGetRepository = r.githubClient.Repositories.GetLatestRelease(ctx, dependency.Owner, dependency.Repo)
	}

	if errGetRepository != nil {
		return nil, errGetRepository //nolint:wrapcheck // TODO
	}

	slog.Info("found repository release: " + githubRelease.GetName())

	switch foundAssets := r.filterAssets(githubRelease.Assets); len(foundAssets) {
	case 1:
		return foundAssets[0], nil

	case 0:
		return nil, errors.New("no github asset found")

	default:
		return nil, errors.New("found more than one github asset")
	}
}

func (r *installer) filterAssets(assets []*github.ReleaseAsset) []*github.ReleaseAsset {
	var foundAssets []*github.ReleaseAsset

	for _, githubAsset := range assets {
		var downloadURL = githubAsset.GetBrowserDownloadURL()

		if strings.Contains(downloadURL, runtime.GOOS+"-"+runtime.GOARCH) {
			foundAssets = append(foundAssets, githubAsset)
		} else {
			var mask string

			switch goos := runtime.GOOS; goos {
			case "windows":
				mask = "win" + runtime.GOARCH[len(runtime.GOARCH)-2:]

			default:
				slog.Warn("TODO: unsupported os: " + goos)
			}

			if strings.Contains(downloadURL, mask) {
				foundAssets = append(foundAssets, githubAsset)
			}
		}
	}

	return foundAssets
}

func (r *installer) downloadAsset(ctx context.Context, asset *github.ReleaseAsset, outFile *os.File) error {
	var (
		getCtx, cancel = context.WithTimeout(ctx, itbasisBuilderHttp.DefaultDownloadTimeout)
		assetURL       = asset.GetBrowserDownloadURL()
	)

	defer cancel()

	slog.Info(fmt.Sprintf("downloading '%s'...", assetURL))

	req, errReq := http.NewRequestWithContext(getCtx, http.MethodGet, assetURL, nil)
	if errReq != nil {
		return errors.WithMessagef(errReq, "fail create request from url: %s", assetURL)
	}

	resp, errResp := r.httpClient.Do(req)
	if errResp != nil {
		return errors.WithMessagef(errResp, "fail download '%s'", assetURL)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if _, err := io.Copy(outFile, resp.Body); err != nil {
		return errors.WithMessagef(err, "fail save downloaded file: %s", assetURL)
	}

	return nil
}

func (r *installer) extractAsset(name model.DependencyName, tmpFile *os.File) error {
	var tmpDirPath, errMkdir = os.MkdirTemp("", name+"-*")
	if errMkdir != nil {
		return errors.WithMessage(errMkdir, "fail create temporary directory")
	}

	defer func() {
		_ = os.RemoveAll(tmpDirPath)
	}()

	if _, _, _, err := xtractr.ExtractFile(&xtractr.XFile{
		FilePath:   tmpFile.Name(),
		OutputDir:  tmpDirPath,
		SquashRoot: true,
		DirMode:    xtractr.DefaultDirMode,
		FileMode:   xtractr.DefaultFileMode,
	}); err != nil {
		return errors.WithMessage(err, "fail extract archive")
	}

	if err := filepath.Walk(tmpDirPath, func(path string, info fs.FileInfo, _ error) error {
		if path == tmpDirPath {
			return nil
		}

		if !info.IsDir() && slices.Contains(excludeFiles, info.Name()) {
			return nil
		}

		var targetPath = filepath.Join(r.binPath, info.Name())
		if runtime.GOOS == "windows" {
			from, _ := syscall.UTF16PtrFromString(path)
			to, _ := syscall.UTF16PtrFromString(targetPath)

			_ = os.Remove(targetPath)

			return syscall.MoveFile(from, to)
		} else {
			return os.Rename(path, targetPath)
		}
	}); err != nil {
		return errors.WithMessage(err, "fail moving unpack files")
	}

	return nil
}
