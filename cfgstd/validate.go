package cfgstd

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"strings"
)

type Validator struct {
	Standard *CfgStandard
}

type ConfigInputEntry struct {
	Key   *yaml.Node
	Value *yaml.Node
}

type ConfigInput []*ConfigInputEntry

func (ci *ConfigInput) UnmarshalYAML(root *yaml.Node) error {
	for i := 0; i < len(root.Content); i += 2 {
		k := root.Content[i]
		v := root.Content[i+1]
		if k.Kind != yaml.ScalarNode {
			return fmt.Errorf("unexpected key type (key index %d, line %d, col %d): %v", i, k.Line, k.Column, k.Kind)
		}
		*ci = append(*ci, &ConfigInputEntry{
			Key:   k,
			Value: v,
		})
	}
	return nil
}

func (val *Validator) Validate(input ConfigInput, errW io.Writer) (valid bool) {
	valid = true
	checkBool := func(e *ConfigInputEntry, b bool, msg string, data ...interface{}) bool {
		if b {
			_, _ = fmt.Fprintf(os.Stderr, "(key: '%s', line: %d): ", e.Key.Value, e.Key.Line)
			_, _ = fmt.Fprintf(os.Stderr, msg, data...)
			_, _ = fmt.Fprintf(os.Stderr, "\n")
			valid = false
			return true
		}
		return false
	}

	var prev *ConfigInputEntry
	for i, e := range input {
		if prev != nil {
			checkBool(e, prev.Key.Line >= e.Key.Line, "entry %d not well separated from previous entry", i)
		}
		prev = e
	}

	type standardAny struct {
		i            int
		fork         string
		constant     *ForkConstant
		configurable *ForkConfigurable
	}

	expected := make(map[string]*standardAny)
	i := 0
	for _, cg := range val.Standard.Constants {
		for _, c := range cg.Entries {
			expected[c.Key] = &standardAny{i: i, fork: cg.Fork, constant: &c}
			i += 1
		}
	}

	for _, cg := range val.Standard.Configurables {
		for _, c := range cg.Entries {
			expected[c.Key] = &standardAny{i: i, fork: cg.Fork, configurable: &c}
			i += 1
		}
	}

	for i, e := range input {
		// everything should be uppercase
		checkBool(e, e.Key.Value != strings.ToUpper(e.Key.Value), "name is not uppercase")
		// names shouldn't be books
		checkBool(e, len(e.Key.Value) > 100, "name is more than 100 chars")

		// consistency with spec
		sp, ok := expected[e.Key.Value]
		if !ok {
			checkBool(e, true, "config entry not recognized")
			continue
		}
		checkBool(e, sp.i != i, "config entry does not match spec configurable (got %d, expected %d, part of fork %s)", i, sp.i, sp.fork)

		var expectedTyp EntryType
		if sp.constant != nil {
			expectedTyp = sp.constant.Typ
		}
		if sp.configurable != nil {
			expectedTyp = sp.configurable.Typ
		}
		// data types check
		checkBool(e, e.Value.Kind != expectedTyp.Kind(), "config value must be %s yaml kind, found kind: %s", fmtKind(expectedTyp.Kind()), fmtKind(e.Value.Kind))
		checkBool(e, e.Value.Style != expectedTyp.Style(), "config values must be %s yaml style, found style: %s", fmtStyle(expectedTyp.Style()), fmtStyle(e.Value.Style))
		checkBool(e, e.Value.Tag != "", "config values must not use yaml tags for typing, as their type is known already and parsing is kept simple. Found tag: %s", e.Value.Tag)

		// No advanced yml node features, keep parsing simple in any environment
		checkBool(e, e.Value.Anchor != "", "config entries should not use the YAML anchor feature, found anchor '%s'", e.Value.Anchor)
		checkBool(e, e.Value.Alias != nil, "config entries should not use the YAML alias feature, alias to node at line %d, col %d, kind %s", e.Value.Alias.Line, e.Value.Alias.Column, fmtKind(e.Value.Kind))
	}
	return
}
