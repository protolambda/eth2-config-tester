package cfgstd

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
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

func (t EntryType) Style() yaml.Style {
	return yaml.LiteralStyle
}

func (t EntryType) InlineValue() bool {
	return true
}

func parseEntryType(v string) (EntryType, error) {
	switch v {
	case "uint":
		return specUint64, nil
	case "bytes1":
		return specBytes1, nil
	case "bytes4":
		return specBytes4, nil
	case "domain":
		return specBLSDomain, nil
	case "bytes32":
		return specBytes32, nil
	case "eth1_addr":
		return specEth1Addr, nil
	case "bignum":
		return specBigNum, nil
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
	Entries []ForkConstant
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
	Entries []ForkConfigurable
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
	Key   string    `yml:"-"`
	Typ   EntryType `yml:"type"`
	Value string    `yml:"value"`
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
		if err := v.Decode(&dat); err != nil {
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
