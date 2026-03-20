// Package main is the entry point for the generic-gantry module.
package main

import (
	"context"

	"go.viam.com/rdk/components/gantry"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/module"
	"go.viam.com/utils"

	"github.com/viam-modules/generic-gantry/multiaxis"
	"github.com/viam-modules/generic-gantry/singleaxis"
)

func main() {
	utils.ContextualMain(mainWithArgs, module.NewLoggerFromArgs("generic-gantry"))
}

func mainWithArgs(ctx context.Context, args []string, logger logging.Logger) error {
	mod, err := module.NewModuleFromArgs(ctx)
	if err != nil {
		return err
	}

	if err = mod.AddModelFromRegistry(ctx, gantry.API, singleaxis.Model); err != nil {
		return err
	}

	if err = mod.AddModelFromRegistry(ctx, gantry.API, multiaxis.Model); err != nil {
		return err
	}

	err = mod.Start(ctx)
	defer mod.Close(ctx)

	if err != nil {
		return err
	}

	<-ctx.Done()

	return nil
}
