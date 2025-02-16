package main_test

// tag::documentation[]
import (
	"context"

	"github.com/itbasis/go-tools-builder/cmd"
)

// If arguments were not passed, they are taken from the `os.Args`.
func SnippetRunWithoutArguments() {
	cmd.InitApp(context.Background()).Run()
}

func SnippetRunWithArguments() {
	cmd.InitApp(context.Background()).Run("generate", "--debug")
}

// end::documentation[]
