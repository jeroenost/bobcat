package test_helpers

import (
	"github.com/ThoughtWorksStudios/datagen/dsl"
	"log"
	"time"
)

func FieldNode(name string, kind dsl.Node, args ...dsl.Node) dsl.Node {
	if len(args) > 0 {
		return dsl.Node{Kind: "field", Name: name, Value: kind, Args: args, Ref: RefInfo("thing.lang", 1, 1, 1)}
	}
	return dsl.Node{Kind: "field", Name: name, Value: kind, Ref: RefInfo("thing.lang", 1, 1, 1)}
}

func RefInfo(filename string, line, col, offset int) *dsl.Location {
	return dsl.NewLocation(filename, line, col, offset)
}

func BuiltinNode(value string) dsl.Node {
	return dsl.Node{Kind: "builtin", Value: value}
}

func StringNode(val string) dsl.Node {
	return dsl.Node{Kind: "literal-string", Value: val}
}

func IntNode(val int64) dsl.Node {
	return dsl.Node{Kind: "literal-int", Value: val}
}

func FloatNode(val float64) dsl.Node {
	return dsl.Node{Kind: "literal-float", Value: val}
}

func DateNode(val string) dsl.Node {
	parsed, err := time.Parse("2006-01-02", val)

	if err != nil {
		log.Fatalf("could not parse %v against YYYY-mm-dd. Error: %v", val, err)
	}

	return dsl.Node{Kind: "literal-date", Value: parsed}
}

func StringArgs(values ...string) []dsl.Node {
	i, size := 0, len(values)
	args := make([]dsl.Node, size)

	for _, val := range values {
		args[i] = StringNode(val)
		i = i + 1
	}

	return args
}

func IntArgs(values ...int64) []dsl.Node {
	i, size := 0, len(values)
	args := make([]dsl.Node, size)

	for _, val := range values {
		args[i] = IntNode(val)
		i = i + 1
	}

	return args
}

func FloatArgs(values ...float64) []dsl.Node {
	i, size := 0, len(values)
	args := make([]dsl.Node, size)

	for _, val := range values {
		args[i] = FloatNode(val)
		i = i + 1
	}

	return args
}

func DateArgs(values ...string) []dsl.Node {
	i, size := 0, len(values)
	args := make([]dsl.Node, size)

	for _, val := range values {
		args[i] = DateNode(val)
		i = i + 1
	}

	return args
}

func RootNode(nodes ...dsl.Node) dsl.Node {
	return dsl.Node{Kind: "root", Children: nodes}
}

func GenerationNode(entityName string, count int64) dsl.Node {
	return dsl.Node{Kind: "generation", Name: entityName, Args: IntArgs(count)}
}

func EntityNode(name string, fields []dsl.Node) dsl.Node {
	return dsl.Node{Name: name, Kind: "definition", Children: fields}
}
