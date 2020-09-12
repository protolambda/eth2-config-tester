package cfgstd

import (
	"fmt"
	"gopkg.in/yaml.v3"
)

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
