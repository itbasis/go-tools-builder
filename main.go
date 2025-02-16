package main

import (
	"context"

	"github.com/itbasis/go-tools-builder/cmd"
)

func main() {
	cmd.InitApp(context.Background()).Run()
}
