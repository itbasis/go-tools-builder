package build

import (
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"runtime"

	builderCmd "github.com/itbasis/go-tools-builder/internal/cmd"
	itbasisBuilderExec "github.com/itbasis/go-tools-builder/internal/exec"
	itbasisCoreCmd "github.com/itbasis/go-tools-core/cmd"
	itbasisCoreEnv "github.com/itbasis/go-tools-core/env"
	itbasisCoreExec "github.com/itbasis/go-tools-core/exec"
	itbasisCoreLog "github.com/itbasis/go-tools-core/log"
	itbasisCoreVersion "github.com/itbasis/go-tools-core/version"
	"github.com/spf13/cobra"
)

var (
	_flagOs      string
	_flagArch    string
	_flagOutput  string
	_flagVersion = itbasisCoreVersion.Unversioned
)

func NewBuildCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:    itbasisCoreCmd.BuildUse("build", builderCmd.UseArgPath),
		Short:  "Building an application for the current platform",
		Args:   cobra.MatchAll(cobra.OnlyValidArgs, cobra.MaximumNArgs(1)),
		PreRun: itbasisCoreCmd.LogCommand,
		Run:    _runBuild,
	}

	cmd.Flags().StringVarP(&_flagOutput, "output", "", "", "")
	cmd.Flags().StringVarP(&_flagOs, "build-os", "", runtime.GOOS, "")
	cmd.Flags().StringVarP(&_flagArch, "build-arch", "", runtime.GOARCH, "")
	cmd.Flags().StringVarP(&_flagVersion, "build-version", "", _flagVersion, "")

	return cmd
}

func _runBuild(cmd *cobra.Command, args []string) {
	var (
		ctx            = cmd.Context()
		versionPkgPath = reflect.TypeFor[itbasisCoreVersion.Version]().PkgPath() + ".version"
		buildArgs      = []string{
			`-trimpath`,
			`-pgo`, `auto`,
			`-ldflags`, `-w -extldflags '-static' -X '` + versionPkgPath + `=` + _flagVersion + `'`,
			`-tags`, `musl`,
		}
	)

	if _flagOutput != "" {
		buildArgs = append(buildArgs, "-o", _flagOutput)

		itbasisCoreCmd.RequireNoError(cmd, os.MkdirAll(filepath.Dir(_flagOutput), os.ModePerm))
	}

	buildArgs = append(buildArgs, args[0])

	slog.Debug("build with arguments", itbasisCoreLog.SlogAttrSliceWithSeparator("buildArgs", " ", buildArgs))

	execGoBuild, errGoBuild := itbasisBuilderExec.NewGoBuildWithCobra(ctx, cmd)
	itbasisCoreCmd.RequireNoError(cmd, errGoBuild)
	itbasisCoreCmd.RequireNoError(
		cmd,
		execGoBuild.Execute(
			ctx,
			itbasisCoreExec.WithRestoreEnv(
				itbasisCoreEnv.MergeEnvs(
					os.Environ(),
					itbasisCoreEnv.Map{
						"GOOS":   _flagOs,
						"GOARCH": _flagArch,
					},
				),
			),
			itbasisCoreExec.WithRestoreArgsIncludePrevious(itbasisCoreExec.IncludePrevArgsBefore, buildArgs...),
		),
	)
}
