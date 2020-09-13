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

	constants := make(map[string]struct{})
	for _, cg := range val.Standard.Constants {
		for _, c := range cg.Entries {
			constants[c.Key] = struct{}{}
		}
	}
	// if a config has constants, it needs to have them all.
	var configHasConstants bool
	for _, e := range input {
		if _, ok := constants[e.Key.Value]; ok {
			configHasConstants = true
			break
		}
	}

	type standardAny struct {
		i            int
		fork         string
		constant     *ForkConstant
		configurable *ForkConfigurable
	}

	expected := make(map[string]*standardAny)
	i := 0
	if configHasConstants {
		for _, cg := range val.Standard.Constants {
			for _, c := range cg.Entries {
				tmp := c
				expected[c.Key] = &standardAny{i: i, fork: cg.Fork, constant: &tmp}
				i += 1
			}
		}
	}

	for _, cg := range val.Standard.Configurables {
		for _, c := range cg.Entries {
			tmp := c
			expected[c.Key] = &standardAny{i: i, fork: cg.Fork, configurable: &tmp}
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
		checkBool(e, sp.i != i, "config entry position does not match spec configurable position (got %d, expected %d, part of fork %s)", i, sp.i, sp.fork)

		var expectedTyp EntryType
		if sp.constant != nil {
			expectedTyp = sp.constant.Typ
		} else if sp.configurable != nil {
			expectedTyp = sp.configurable.Typ
		} else {
			fmt.Println("missing typ")
		}
		// data types check
		checkBool(e, e.Value.Kind != expectedTyp.Kind(), "config value must be %s yaml kind, found kind: %s", fmtKind(expectedTyp.Kind()), fmtKind(e.Value.Kind))
		checkBool(e, e.Value.Style != expectedTyp.Style(), "config values must be %s yaml style, found style: %s", fmtStyle(expectedTyp.Style()), fmtStyle(e.Value.Style))
		if expectedTyp.Style()&yaml.TaggedStyle != 0 {
			checkBool(e, e.Value.Style&yaml.TaggedStyle == 0, "config value must have explicit tag %s, but got none", expectedTyp.Tag())
			checkBool(e, e.Value.Tag != expectedTyp.Tag(), "config value must match tag %s, got: %s", expectedTyp.Tag(), e.Value.Tag)
		} else {
			checkBool(e, e.Value.Style&yaml.TaggedStyle != 0, "config value must not have an explicit tag, but got", e.Value.Tag)
		}
		if e.Value.Kind == yaml.ScalarNode {
			vErr := expectedTyp.CheckFormatting(e.Value.Value)
			checkBool(e, vErr != nil, "config value has bad formatting: %v", vErr)
		} else {

		}

		// No advanced yml node features, keep parsing simple in any environment
		checkBool(e, e.Value.Anchor != "", "config entries should not use the YAML anchor feature, found anchor '%s'", e.Value.Anchor)
		checkBool(e, e.Value.Alias != nil, "config entries should not use the YAML alias feature, alias to node: ", fmtUnknownNode(e.Value.Alias))
	}
	return
}
