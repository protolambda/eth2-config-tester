package cfgstd

import (
	"encoding/hex"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"strconv"
	"strings"
)

type EntryType int

const (
	specUint64 = iota
	specBytes1
	specBytes4
	specBLSDomain
	specBytes32
	specEth1Addr
	specBigNum
	specOffsets
)

func (t EntryType) Kind() yaml.Kind {
	switch t {
	case specOffsets:
		return yaml.SequenceNode
	default:
		return yaml.ScalarNode
	}
}

// Tag is only checked if the EntryType requires a yaml.TaggedStyle bit set in its Style()
func (t EntryType) Tag() string {
	switch t {
	case specEth1Addr:
		return "!!str"
	case specOffsets:
		return "!!seq"
	default:
		return "!!int"
	}
}

func checkHex(v string, expectedByteLen int, mustLowercase bool) error {
	if !strings.HasPrefix(v, "0x") {
		return fmt.Errorf("missing 0x prefix in value: %s", v)
	}
	dat := v[2:]
	if len(dat) != expectedByteLen*2 {
		return fmt.Errorf("expected %d bytes (%d hex chars), but got %d hex chars: %s", expectedByteLen, expectedByteLen*2, len(dat), v)
	}
	if _, err := hex.DecodeString(dat); err != nil {
		return fmt.Errorf("invalid hex string: %v", err)
	}
	if mustLowercase && strings.ToLower(v) != v {
		return fmt.Errorf("expected lowercase hex string, but got: %s", v)
	}
	return nil
}

func (t EntryType) CheckFormatting(v string) error {
	switch t {
	case specUint64:
		_, err := strconv.ParseUint(v, 10, 64)
		return err
	case specBytes1:
		return checkHex(v, 1, true)
	case specBytes4, specBLSDomain:
		return checkHex(v, 4, true)
	case specBytes32:
		return checkHex(v, 4, true)
	case specEth1Addr:
		return checkHex(v, 20, false) // TODO check eth1 address checksum?
	case specBigNum:
		for _, c := range v {
			if c < '0' || c > '9' {
				return fmt.Errorf("invalid bignum decimal number: %s", v)
			}
		}
		return nil
	case specOffsets:
		return fmt.Errorf("cannot check formatting of non-scalar types")
	default:
		return fmt.Errorf("unrecognized entry type %d, cannot check formatting", t)
	}
}

func (t EntryType) CheckContents(contents []*yaml.Node) error {
	switch t {
	case specOffsets:
		p := uint64(0)
		for i, d := range contents {
			if d.Kind != yaml.ScalarNode {
				return fmt.Errorf("offset %d should be scalar node", i)
			}
			if d.Style != 0 {
				return fmt.Errorf("offset %d should not be styled", i)
			}
			x, err := strconv.ParseUint(d.Value, 10, 64)
			if err != nil {
				return fmt.Errorf("offset %d is invalid uint64: %v", i, err)
			}
			if x < p {
				return fmt.Errorf("offset %d decreases: %d -> %d", p, x)
			}
			p = x
		}
		return nil
	default:
		return nil
	}
}

func (t EntryType) Style() yaml.Style {
	switch t {
	case specOffsets:
		return yaml.FlowStyle
	default:
		// no style
		return 0
	}
}

func parseEntryType(v string) (EntryType, error) {
	switch v {
	case "uint64":
		return specUint64, nil
	case "bytes1":
		return specBytes1, nil
	case "bytes4":
		return specBytes4, nil
	case "domain":
		return specBLSDomain, nil
	case "bytes32":
		return specBytes32, nil
	case "eth1_address":
		return specEth1Addr, nil
	case "bignum":
		return specBigNum, nil
	case "offsets":
		return specOffsets, nil
	default:
		return 0, fmt.Errorf("unrecognized entry type: %s", v)
	}
}

func (t *EntryType) UnmarshalYAML(root *yaml.Node) error {
	p, err := parseEntryType(root.Value)
	if err != nil {
		return err
	}
	*t = p
	return nil
}

type ConstantsForkGroup struct {
	Fork    string
	Entries ConstantsForkSeq
}

// ordered map, fork name -> constants of fork
type ConstantsByFork []ConstantsForkGroup

func (c *ConstantsByFork) UnmarshalYAML(root *yaml.Node) error {
	for i := 0; i < len(root.Content); i += 2 {
		k := root.Content[i]
		v := root.Content[i+1]
		if k.Kind != yaml.ScalarNode {
			return fmt.Errorf("unexpected key type: %v", k.Kind)
		}
		var dat ConstantsForkSeq
		if err := v.Decode(&dat); err != nil {
			return fmt.Errorf("failed to decode '%s': %v", k.Value, err)
		}
		*c = append(*c, ConstantsForkGroup{Fork: k.Value, Entries: dat})
	}
	return nil
}

type ForkConstant struct {
	Key   string    `yml:"-"`
	Typ   EntryType `yml:"type"`
	Value string    `yml:"value"`
}

// ordered map, constant name -> constant spec
type ConstantsForkSeq []ForkConstant

func (c *ConstantsForkSeq) UnmarshalYAML(root *yaml.Node) error {
	for i := 0; i < len(root.Content); i += 2 {
		k := root.Content[i]
		v := root.Content[i+1]
		if k.Kind != yaml.ScalarNode {
			return fmt.Errorf("unexpected key type: %v", k.Kind)
		}
		var dat ForkConstant
		if err := v.Decode(&dat); err != nil {
			return fmt.Errorf("failed to decode '%s': %v", k.Value, err)
		}
		dat.Key = k.Value
		*c = append(*c, dat)
	}
	return nil
}

type ConfigurablesForkGroup struct {
	Fork    string
	Entries ConfigurablesForkSeq
}

// ordered map, fork name -> configurables of fork
type ConfigurablesByFork []ConfigurablesForkGroup

func (c *ConfigurablesByFork) UnmarshalYAML(root *yaml.Node) error {
	for i := 0; i < len(root.Content); i += 2 {
		k := root.Content[i]
		v := root.Content[i+1]
		if k.Kind != yaml.ScalarNode {
			return fmt.Errorf("unexpected key type: %v", k.Kind)
		}
		var dat ConfigurablesForkSeq
		if err := v.Decode(&dat); err != nil {
			return fmt.Errorf("failed to decode '%s': %v", k.Value, err)
		}
		*c = append(*c, ConfigurablesForkGroup{Fork: k.Value, Entries: dat})
	}
	return nil
}

type ForkConfigurable struct {
	Key string
	Typ EntryType
}

// ordered map, configurable name -> configurable spec
type ConfigurablesForkSeq []ForkConfigurable

func (c *ConfigurablesForkSeq) UnmarshalYAML(root *yaml.Node) error {
	for i := 0; i < len(root.Content); i += 2 {
		k := root.Content[i]
		v := root.Content[i+1]
		if k.Kind != yaml.ScalarNode {
			return fmt.Errorf("unexpected key type: %v", k.Kind)
		}
		var dat ForkConfigurable
		if err := v.Decode(&dat.Typ); err != nil {
			return fmt.Errorf("failed to decode '%s': %v", k.Value, err)
		}
		dat.Key = k.Value
		*c = append(*c, dat)
	}
	return nil
}

type CfgStandard struct {
	Constants     ConstantsByFork     `yaml:"constants"`
	Configurables ConfigurablesByFork `yaml:"configurables"`
}

func (cfgStd *CfgStandard) ValidateSelf() error {
	type entry struct {
		category string
		name     string
		fork     string
	}
	byName := make(map[string]entry)
	for _, g := range cfgStd.Constants {
		for _, e := range g.Entries {
			if p, ok := byName[e.Key]; ok {
				return fmt.Errorf("constant %s in fork %s conflicts with %s %s in fork %s", e.Key, g.Fork, p.category, p.name, p.fork)
			}
			byName[e.Key] = entry{category: "constant", name: e.Key, fork: g.Fork}
		}
	}
	for _, g := range cfgStd.Configurables {
		for _, e := range g.Entries {
			if p, ok := byName[e.Key]; ok {
				return fmt.Errorf("configurable %s in fork %s conflicts with %s %s in fork %s", e.Key, g.Fork, p.category, p.name, p.fork)
			}
			byName[e.Key] = entry{category: "constant", name: e.Key, fork: g.Fork}
		}
	}
	return nil
}

func LoadStandard(path string) (out *CfgStandard, err error) {
	var cfgStd CfgStandard
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("config standard, cannot oppen file: %v", err)
	}
	defer f.Close()
	dec := yaml.NewDecoder(f)
	if err := dec.Decode(&cfgStd); err != nil {
		return nil, fmt.Errorf("config standard, decoding error: %v", err)
	}
	if err := cfgStd.ValidateSelf(); err != nil {
		return nil, fmt.Errorf("loaded invalid standard: %v", err)
	}
	return &cfgStd, nil
}
