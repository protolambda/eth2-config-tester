package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
)

func checkErr(msg string, err error) {
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s: %v\n", msg, err)
	}
}

func checkBool(k string, v *yaml.Node, b bool, msg string, data ...interface{}) {
	if b {
		_, _ = fmt.Fprintf(os.Stderr, "(key: '%s', line: %d): ", k, v.Line)
		_, _ = fmt.Fprintf(os.Stderr, msg, data...)
		_, _ = fmt.Fprintf(os.Stderr, "\n")
	}
}

// Format YAML kinds. Think of it as the node structure type.
func fmtKind(k yaml.Kind) string {
	switch k {
	case yaml.DocumentNode:
		return "document"
	case yaml.SequenceNode:
		return "sequence"
	case yaml.MappingNode:
		return "mapping"
	case yaml.ScalarNode:
		return "scalar"
	case yaml.AliasNode:
		return "alias"
	default:
		return fmt.Sprintf("unrecognized! (%d)", int(k))
	}
}

// Format YAML styles. Think of it as the representation of the node, does not include integer formatting etc.
func fmtStyle(s yaml.Style) string {
	switch s {
	case yaml.TaggedStyle: // a.k.a. plain style
		return "tagged"
	case yaml.DoubleQuotedStyle:
		return "double-quotes"
	case yaml.SingleQuotedStyle:
		return "single-quotes"
	case yaml.LiteralStyle:
		return "literal"
	case yaml.FoldedStyle:
		return "folded"
	case yaml.FlowStyle:
		return "flow"
	default:
		return fmt.Sprintf("unrecognized! (%d)", int(s))
	}
}

func main() {
	var cfg map[string]*yaml.Node
	dec := yaml.NewDecoder(os.Stdin)
	if err := dec.Decode(&cfg); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to even decode config: %v", err)
		os.Exit(1)
	}

	// naming consistency
	for k, v := range cfg {
		// everything should be uppercase
		checkBool(k, v, k != strings.ToUpper(k), "name is not uppercase")
		// names shouldn't be books
		checkBool(k, v, len(k) > 100, "name is more than 100 chars")
	}

	// data types check
	for k, v := range cfg {
		// We can add exceptions for some keys, but generally the formatting is simple and direct literals
		checkBool(k, v, v.Kind != yaml.ScalarNode, "config values must be scalar nodes, found kind: %s", fmtKind(v.Kind))
		checkBool(k, v, v.Style != yaml.LiteralStyle, "config values must not be quoted or styled otherwise, found style: %s", fmtKind(v.Kind))
		checkBool(k, v, v.Tag != "", "config values must not use yaml tags for typing, as their type is known already and parsing is kept simple. Found tag: %s", v.Tag)
	}

	// No advanced yml node features, keep parsing simple in any environment
	for k, v := range cfg {
		checkBool(k, v, v.Anchor != "", "config entries should not use the YAML anchor feature, found anchor '%s'", v.Anchor)
		checkBool(k, v, v.Alias != nil, "config entries should not use the YAML alias feature, alias to node at line %d, col %d, kind %s", v.Alias.Line, v.Alias.Column, fmtKind(v.Kind))
	}

}
