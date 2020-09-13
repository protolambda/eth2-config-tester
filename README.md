# Eth2 config tester

The goals is to make configs:
- Simple to parse, key-value
- Valid YAML, but no advanced features
- Compatible, no unexpected or missing keys
- Consistent, no formatting differences
- Minimal, no super long names 
- Familiar, use common (within ethereum) 0x prefix for bytes
- Structured, prefer a specific order of config keys
- Bare, no inline comments, head/foot comments are ok.
- Standard, following a spec for the config

## Usage

```shell script
# install
go get github.com/protolambda/eth2-config-tester

# feed config file into tester
cat my_eth2_config.yml | eth2-config-tester --config-spec=config_spec.yaml
```

## Planned

- Check config value correctness, bounds etc. See https://github.com/ethereum/eth2.0-specs/issues/407
- After checking for all bad config values, output a best-effort properly formatted config


## License

CC0 1.0 Universal, see [LICENSE](./LICENSE) file.
