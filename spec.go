package main

import (
	"fmt"
	"os"
	"encoding/json"

	"github.com/urfave/cli"
	"github.com/opencontainers/runtime-spec/specs-go"
)

// loadSpec loads the specification from the provided path.
func loadSpec(cPath string) (spec *specs.Spec, err error) {
	cf, err := os.Open(cPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("JSON specification file %s not found", cPath)
		}
		return nil, err
	}
	defer cf.Close()

	if err = json.NewDecoder(cf).Decode(&spec); err != nil {
		return nil, err
	}
	return spec, nil
}

// setupSpec performs initial setup based on the cli.Context for the container
func setupSpec(context *cli.Context) (*specs.Spec, error) {
	bundle := context.String("bundle")
	if bundle != "" {
		if err := os.Chdir(bundle); err != nil {
			fmt.Printf("error: dir not found (%s)\n", bundle)
			return nil, err
		}
	}
	spec, err := loadSpec(specConfig)
	if err != nil {
		fmt.Printf("loadSpec err (%s)\n", err)
		return nil, err
	}
	return spec, nil
}
