package main

import (
	"eth2-config-tester/cfgstd"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

func main() {
	cfgStdPath := "config_spec.yaml"
	cfgStd, err := cfgstd.LoadStandard(cfgStdPath)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to load config standard '%s': %v\n", cfgStdPath, err)
	}
	var cfg cfgstd.ConfigInput
	dec := yaml.NewDecoder(os.Stdin)
	if err := dec.Decode(&cfg); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to even decode config: %v", err)
		os.Exit(1)
	}

	validator := cfgstd.Validator{Standard: cfgStd}
	if validator.Validate(cfg, os.Stderr) {
		os.Exit(0)
	} else {
		os.Exit(1)
	}
}
