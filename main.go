package main

import (
	"eth2-config-tester/cfgstd"
	"flag"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

func main() {
	cfgStdPath := flag.String("config-spec", "config_spec.yaml", "configuration standard file to validate configs with")
	flag.Parse()

	cfgStd, err := cfgstd.LoadStandard(*cfgStdPath)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to load config standard '%s': %v\n", *cfgStdPath, err)
		os.Exit(1)
	}
	var cfg cfgstd.ConfigInput
	dec := yaml.NewDecoder(os.Stdin)
	if err := dec.Decode(&cfg); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to even decode config: %v", err)
		os.Exit(1)
	}

	validator := cfgstd.Validator{Standard: cfgStd}
	if validator.Validate(cfg, os.Stderr) {
		_, _ = fmt.Fprintf(os.Stderr, "config is valid! (config spec: %s)\n", *cfgStdPath)
		os.Exit(0)
	} else {
		os.Exit(1)
	}
}
