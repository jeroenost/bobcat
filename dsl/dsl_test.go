package dsl

import (
	"fmt"
	. "github.com/ThoughtWorksStudios/bobcat/common"
	. "github.com/ThoughtWorksStudios/bobcat/test_helpers"
	re "regexp"
	"testing"
)

func runParser(script string, tokens ...interface{}) (interface{}, error) {
	if nil != tokens && len(tokens) > 0 {
		script = fmt.Sprintf(script, tokens...)
	}

	return Parse("testScript", []byte(script), Recover(false))
}

func testEntity(name, parent string, fields NodeSet) *Node {
	var parentNode *Node

	if "" != parent {
		parentNode = IdNode(nil, parent)
	}

	var body *Node

	if len(fields) > 0 {
		body = EntityBodyNode(nil, nil, FieldSetNode(nil, fields))
	} else {
		body = EntityBodyNode(nil, nil, nil)
	}

	if "" != name {
		return EntityNode(nil, IdNode(nil, name), parentNode, body)
	} else {
		return EntityNode(nil, nil, parentNode, body)
	}
}

func removeLocationInfo(err error) error {
	if nil == err {
		return nil
	}

	prefix := re.MustCompile(`^testScript:\d+:\d+ \(\d+\):\s+(?:rule (?:"[\w -]+"|\w+):\s+)?`)
	return fmt.Errorf(prefix.ReplaceAllString(err.Error(), ""))
}

func TestBuiltinsAsExpressions(t *testing.T) {
	actual, err := runParser(`
	  $int(1,4)
	`)
	AssertNil(t, err, "Didn't expect to get an error: %v", err)

	expected := RootNode(nil, NodeSet{
		CallNode(nil,
			BuiltinNode(nil, INT_TYPE),
			NodeSet{IntLiteralNode(nil, 1), IntLiteralNode(nil, 4)},
		),
	})
	AssertEqual(t, expected.String(), actual.(*Node).String())
}

func TestParsesBasicEntity(t *testing.T) {
	testRoot := RootNode(nil, NodeSet{testEntity("Bird", "", NodeSet{})})
	actual, err := runParser("entity Bird {  }")
	AssertNil(t, err, "Didn't expect to get an error: %v", err)
	AssertEqual(t, testRoot.String(), actual.(*Node).String())
}

func TestParseAnonymousEntity(t *testing.T) {
	testRoot := RootNode(nil, NodeSet{testEntity("", "", NodeSet{})})
	actual, err := runParser("entity {  }")
	AssertNil(t, err, "Didn't expect to get an error: %v", err)
	AssertEqual(t, testRoot.String(), actual.(*Node).String())
}

func TestParseBinaryExpression(t *testing.T) {
	expected := RootNode(nil, NodeSet{
		&Node{
			Name: "+", Kind: "binary",
			Value: AtomicNode(nil, IntLiteralNode(nil, 1)), Related: &Node{
				Name: "+", Kind: "binary",
				Value: IntLiteralNode(nil, 2), Related: IntLiteralNode(nil, 3),
			},
		},
	})

	actual, err := runParser(`1 + 2 + 3`)

	AssertNil(t, err, "Didn't expect to get an error: %v", err)
	AssertEqual(t, expected.String(), actual.(*Node).String())
}

func TestBinaryExpressionOperatorPrecedence(t *testing.T) {
	expected := RootNode(nil, NodeSet{
		&Node{
			Name: "+", Kind: "binary",
			Value: &Node{
				Name: "*", Kind: "binary",
				Value: AtomicNode(nil, IntLiteralNode(nil, 1)), Related: IntLiteralNode(nil, 2),
			}, Related: IntLiteralNode(nil, 3),
		},
	})

	actual, err := runParser(`1 * 2 + 3`)
	AssertNil(t, err, "Didn't expect to get an error: %v", err)
	AssertEqual(t, expected.String(), actual.(*Node).String())

	expected = RootNode(nil, NodeSet{
		&Node{
			Name: "+", Kind: "binary",
			Value: AtomicNode(nil, IntLiteralNode(nil, 1)), Related: &Node{
				Name: "*", Kind: "binary",
				Value: IntLiteralNode(nil, 2), Related: IntLiteralNode(nil, 3),
			},
		},
	})

	actual, err = runParser(`1 + 2 * 3`)

	AssertNil(t, err, "Didn't expect to get an error: %v", err)
	AssertEqual(t, expected.String(), actual.(*Node).String())
}

func TestCanParseMultipleEntities(t *testing.T) {
	bird1 := testEntity("Bird", "", NodeSet{})
	bird2 := testEntity("Bird2", "", NodeSet{})
	testRoot := RootNode(nil, NodeSet{bird1, bird2})
	actual, err := runParser("entity Bird {  }\nentity Bird2 { }")
	AssertNil(t, err, "Didn't expect to get an error: %v", err)
	AssertEqual(t, testRoot.String(), actual.(*Node).String())
}

func TestParsesChildEntity(t *testing.T) {
	entity := testEntity("Robin", "", NodeSet{})
	entity.Related = IdNode(nil, "Bird")
	testRoot := RootNode(nil, NodeSet{entity})
	actual, err := runParser("entity Robin << Bird {  }")
	AssertNil(t, err, "Didn't expect to get an error: %v", err)
	AssertEqual(t, testRoot.String(), actual.(*Node).String())
}

func TestParsesBasicGenerationStatement(t *testing.T) {
	args := NodeSet{IntLiteralNode(nil, 1), IdNode(nil, "Bird")}
	genBird := GenNode(nil, args)
	testRoot := RootNode(nil, NodeSet{genBird})
	actual, err := runParser("generate(1, Bird)")
	AssertNil(t, err, "Didn't expect to get an error: %v", err)
	AssertEqual(t, testRoot.String(), actual.(*Node).String())
}

func TestCanParseMultipleGenerationStatements(t *testing.T) {
	arg := IntLiteralNode(nil, 1)
	genBird := GenNode(nil, NodeSet{arg, IdNode(nil, "Bird")})
	bird2Gen := GenNode(nil, NodeSet{arg, IdNode(nil, "Bird2")})
	testRoot := RootNode(nil, NodeSet{genBird, bird2Gen})

	actual, err := runParser("generate(1, Bird)\ngenerate(1, Bird2)")
	AssertNil(t, err, "Didn't expect to get an error: %v", err)
	AssertEqual(t, testRoot.String(), actual.(*Node).String())
}

func TestCanOverrideFieldInGenerateStatement(t *testing.T) {
	arg := IntLiteralNode(nil, 1)
	value := StrLiteralNode(nil, "birdie")
	field := FieldNode(nil, IdNode(nil, "name"), value, nil)
	genBird := GenNode(nil, NodeSet{arg, testEntity("", "Bird", NodeSet{field})})
	testRoot := RootNode(nil, NodeSet{genBird})
	actual, err := runParser("generate(1, Bird << { name: \"birdie\" })")
	AssertNil(t, err, "Didn't expect to get an error: %v", err)
	AssertEqual(t, testRoot.String(), actual.(*Node).String())
}

func TestCanOverrideMultipleFieldsInGenerateStatement(t *testing.T) {
	value1 := StrLiteralNode(nil, "birdie")
	field1 := FieldNode(nil, IdNode(nil, "name"), value1, nil)
	arg1 := IntLiteralNode(nil, 1)
	arg2 := IntLiteralNode(nil, 2)
	value2 := CallNode(nil, BuiltinNode(nil, INT_TYPE), NodeSet{arg1, arg2})
	field2 := FieldNode(nil, IdNode(nil, "age"), value2, nil)

	arg := IntLiteralNode(nil, 1)
	genBird := GenNode(nil, NodeSet{arg, testEntity("", "Bird", NodeSet{field1, field2})})
	testRoot := RootNode(nil, NodeSet{genBird})
	actual, err := runParser("generate(1, Bird << { name: \"birdie\", age: $int(1,2) })")
	AssertNil(t, err, "Didn't expect to get an error: %v", err)
	AssertEqual(t, testRoot.String(), actual.(*Node).String())
}

func TestParsedBothBasicEntityAndGenerationStatement(t *testing.T) {
	args := NodeSet{IntLiteralNode(nil, 1), IdNode(nil, "Bird")}
	genBird := GenNode(nil, args)
	bird := testEntity("Bird", "", NodeSet{})
	testRoot := RootNode(nil, NodeSet{bird, genBird})
	actual, err := runParser("entity Bird {}\ngenerate (1, Bird)")
	AssertNil(t, err, "Didn't expect to get an error: %v", err)
	AssertEqual(t, testRoot.String(), actual.(*Node).String())
}

func TestParseEntityWithExpressionFieldWithBound(t *testing.T) {
	value := CallNode(nil, BuiltinNode(nil, STRING_TYPE), NodeSet{})
	count := RangeNode(nil, IntLiteralNode(nil, 1), IntLiteralNode(nil, 8))

	field := FieldNode(nil, IdNode(nil, "name"), value, count)
	bird := testEntity("Bird", "", NodeSet{field})
	testRoot := RootNode(nil, NodeSet{bird})
	actual, err := runParser("entity Bird { name: $str()<1..8> }")
	AssertNil(t, err, "Didn't expect to get an error: %v", err)
	AssertEqual(t, testRoot.String(), actual.(*Node).String())
}

func TestParseEntityWithExpressionFieldWithoutArgs(t *testing.T) {
	value := CallNode(nil, BuiltinNode(nil, STRING_TYPE), NodeSet{})
	field := FieldNode(nil, IdNode(nil, "name"), value, nil)

	bird := testEntity("Bird", "", NodeSet{field})
	testRoot := RootNode(nil, NodeSet{bird})
	actual, err := runParser("entity Bird { name: $str() }")
	AssertNil(t, err, "Didn't expect to get an error: %v", err)
	AssertEqual(t, testRoot.String(), actual.(*Node).String())
}

func TestParseEntityWithNormalDistributionField(t *testing.T) {
	args := NodeSet{IntLiteralNode(nil, 1), IntLiteralNode(nil, 10)}

	field := FieldNode(nil, IdNode(nil, "age"), DistributionNode(nil, NORMAL_DIST, args), nil)
	bird := testEntity("Bird", "", NodeSet{field})

	expected := RootNode(nil, NodeSet{bird})
	actual, err := runParser("entity Bird { age: %v ~ [1, 10] }", NORMAL_DIST)

	AssertNil(t, err, "Didn't expect to get an error: %v", err)
	AssertEqual(t, expected.String(), actual.(*Node).String())
}

func TestParseEntityWithDistributedStaticField(t *testing.T) {
	intervals := NodeSet{AssociativeArgumentNode(nil, IntLiteralNode(nil, 10), StrLiteralNode(nil, "blah"))}

	field := FieldNode(nil, IdNode(nil, "age"), DistributionNode(nil, PERCENT_DIST, intervals), nil)
	bird := testEntity("Bird", "", NodeSet{field})

	expected := RootNode(nil, NodeSet{bird})
	actual, err := runParser("entity Bird { age: %v ~ [10 => \"blah\"] }", PERCENT_DIST)

	AssertNil(t, err, "Didn't expect to get an error: %v", err)
	AssertEqual(t, expected.String(), actual.(*Node).String())
}

func TestCannotUseDistributionsAsArgs(t *testing.T) {
	_, err1 := runParser("entity Bird { age: %v ~ [ 10 => %v ~ [1,2] ] }", PERCENT_DIST, NORMAL_DIST)
	ExpectsError(t, "Distributions cannot be used as arguments", err1)

	_, err2 := runParser("entity Bird { age: %v ~ [ %v ~ [1,2] => 10 ] }", PERCENT_DIST, NORMAL_DIST)
	ExpectsError(t, "Distributions cannot be used as arguments", err2)

	_, err3 := runParser("entity Bird { age: %v ~ [ %v ~ [1,2] ] }", NORMAL_DIST, NORMAL_DIST)
	ExpectsError(t, "Distributions cannot be used as arguments", err3)

	_, err4 := runParser("(lambda noop(x) { x })(%v ~ [1,2]) }", NORMAL_DIST)
	ExpectsError(t, "Distributions cannot be used as arguments", err4)
}

func TestParseEntityWithUnSupportedDistributionTypeShouldError(t *testing.T) {
	_, err := runParser("entity Bird { age: *foobar ~ [1, 2] }")
	ExpectsError(t, "Unknown distribution \"*foobar\"", err)
}

func TestUnterminatedDistribution(t *testing.T) {
	_, err := runParser("entity Bird { age: %v ~ [1, 2 }", NORMAL_DIST)
	ExpectsError(t, "Unterminated distribution", err)
}

func TestParseEntityWithExpressionFieldWithArgs(t *testing.T) {
	args := NodeSet{IntLiteralNode(nil, 1)}
	value := CallNode(nil, BuiltinNode(nil, STRING_TYPE), args)
	field := FieldNode(nil, IdNode(nil, "name"), value, nil)
	bird := testEntity("Bird", "", NodeSet{field})
	testRoot := RootNode(nil, NodeSet{bird})
	actual, err := runParser("entity Bird { name: $str(1) }")
	AssertNil(t, err, "Didn't expect to get an error: %v", err)
	AssertEqual(t, testRoot.String(), actual.(*Node).String())
}

func TestParseEntityWithMultipleFields(t *testing.T) {
	arg := IntLiteralNode(nil, 1)
	value := CallNode(nil, BuiltinNode(nil, STRING_TYPE), NodeSet{arg})
	field1 := FieldNode(nil, IdNode(nil, "name"), value, nil)

	arg1 := IntLiteralNode(nil, 1)
	arg2 := IntLiteralNode(nil, 5)
	args := NodeSet{arg1, arg2}
	value = CallNode(nil, BuiltinNode(nil, INT_TYPE), args)
	field2 := FieldNode(nil, IdNode(nil, "age"), value, nil)

	bird := testEntity("Bird", "", NodeSet{field1, field2})
	testRoot := RootNode(nil, NodeSet{bird})
	actual, err := runParser("entity Bird { name: $str(1), age: $int(1, 5) }")
	AssertNil(t, err, "Didn't expect to get an error: %v", err)
	AssertEqual(t, testRoot.String(), actual.(*Node).String())
}

func TestParseEntityWithStaticField(t *testing.T) {
	value := StrLiteralNode(nil, "birdie")
	field := FieldNode(nil, IdNode(nil, "name"), value, nil)
	bird := testEntity("Bird", "", NodeSet{field})
	testRoot := RootNode(nil, NodeSet{bird})
	actual, err := runParser("entity Bird { name: \"birdie\" }")
	AssertNil(t, err, "Didn't expect to get an error: %v", err)
	AssertEqual(t, testRoot.String(), actual.(*Node).String())
}

func TestParseEntityWithEntityDeclarationField(t *testing.T) {
	goatValue := StrLiteralNode(nil, "billy")
	goatField := FieldNode(nil, IdNode(nil, "name"), goatValue, nil)
	goat := testEntity("Goat", "", NodeSet{goatField})
	field := FieldNode(nil, IdNode(nil, "pet"), goat, nil)
	person := testEntity("Person", "", NodeSet{field})
	testRoot := RootNode(nil, NodeSet{person})
	actual, err := runParser("entity Person { pet: entity Goat { name: \"billy\" } }")
	AssertNil(t, err, "Didn't expect to get an error: %v", err)
	AssertEqual(t, testRoot.String(), actual.(*Node).String())
}

func TestParseEntityWithEntityReferenceField(t *testing.T) {
	goatValue := StrLiteralNode(nil, "billy")
	goatField := FieldNode(nil, IdNode(nil, "name"), goatValue, nil)
	goat := testEntity("Goat", "", NodeSet{goatField})
	value := IdNode(nil, "Goat")
	field := FieldNode(nil, IdNode(nil, "pet"), value, nil)
	person := testEntity("Person", "", NodeSet{field})
	testRoot := RootNode(nil, NodeSet{goat, person})
	actual, err := runParser("entity Goat { name: \"billy\" } entity Person { pet: Goat }")
	AssertNil(t, err, "Didn't expect to get an error: %v", err)
	AssertEqual(t, testRoot.String(), actual.(*Node).String())
}

func TestVariableAssignment(t *testing.T) {
	value := StrLiteralNode(nil, "hello")
	name := FieldNode(nil, IdNode(nil, "name"), value, nil)
	foo := testEntity("Foo", "", NodeSet{name})
	testRoot := RootNode(nil, NodeSet{
		foo,
		AssignNode(nil,
			IdNode(nil, "foos"),
			GenNode(nil, NodeSet{IntLiteralNode(nil, 3), IdNode(nil, "Foo")}),
		),
	})
	actual, err := runParser("entity Foo { name: \"hello\" } foos = generate(3, Foo)")
	AssertNil(t, err, "Didn't expect to get an error: %v", err)
	AssertEqual(t, testRoot.String(), actual.(*Node).String())
}

func TestValidPrimaryKeys(t *testing.T) {
	for _, stmt := range []string{
		`pk("id", $uid)`,
		`pk("id", $uniqint)`,
		`pk("id", $incr)`,
	} {
		_, err := runParser(stmt)
		AssertNil(t, err, "Should not receive error when parsing %q", stmt)
	}

	_, err := runParser(`pk("id", $float)`)
	errorMessage := fmt.Sprintf("Primary key may only be of type `%s`, `%s`, or `%s`.", SERIAL_TYPE, UNIQUE_INT_TYPE, UID_TYPE)

	ExpectsError(t, errorMessage, err)
}

func TestRequiresValidStatements(t *testing.T) {
	_, err := runParser("!")
	expectedErrorMsg := `Don't know how to evaluate "!"`
	ExpectsError(t, expectedErrorMsg, removeLocationInfo(err))
}

func TestGenerateWithNoArgumentsProducesError(t *testing.T) {
	expectedErrMessage := "`generate` statement \"generate Blah\" requires arguments `(count, entity)`"
	_, err := runParser("generate Blah")
	ExpectsError(t, expectedErrMessage, removeLocationInfo(err))
}

func TestEntityFieldRequiresType(t *testing.T) {
	expectedErrMessage := `Missing field type for field declaration "name"`
	_, err := runParser("entity Blah { name: }")
	ExpectsError(t, expectedErrMessage, removeLocationInfo(err))
}

func TestAssignmentMissingRightHandSide(t *testing.T) {
	expectedErrMessage := `Missing right-hand of assignment expression "Bird ="`
	_, err := runParser("Bird =")
	ExpectsError(t, expectedErrMessage, removeLocationInfo(err))
}

func TestEntityDefinitionRequiresCurlyBrackets(t *testing.T) {
	expectedErrMessage := `Unterminated entity expression (missing closing curly brace)`
	_, err := runParser("entity Bird {")
	ExpectsError(t, expectedErrMessage, removeLocationInfo(err))
}

func TestFieldListWithoutCommas(t *testing.T) {
	expectedErrMessage := `Multiple field declarations must be delimited with a comma`
	_, err := runParser("entity Bird { h: string b: string }")
	ExpectsError(t, expectedErrMessage, removeLocationInfo(err))
}

func TestIllegalIdentifiers(t *testing.T) {
	specs := map[string]string{
		"let 2hot = true":           `Illegal identifier "2hot"; identifiers start with a letter or underscore, followed by zero or more letters, underscores, and numbers`,
		"1luv = [1, 2]":             `Illegal identifier "1luv"; identifiers start with a letter or underscore, followed by zero or more letters, underscores, and numbers`,
		"entity a {$ok: 0}":         `Illegal identifier "$ok"; identifiers start with a letter or underscore, followed by zero or more letters, underscores, and numbers`,
		"entity 4fun { }":           `Illegal identifier "4fun"; identifiers start with a letter or underscore, followed by zero or more letters, underscores, and numbers`,
		"entity $eek { }":           `Illegal identifier "$eek"; identifiers start with a letter or underscore, followed by zero or more letters, underscores, and numbers`,
		"generate (1, $a)":          `Illegal identifier "$a"; identifiers start with a letter or underscore, followed by zero or more letters, underscores, and numbers`,
		"entity generate { }":       `Illegal identifier "generate"; reserved words cannot be used as identifiers`,
		"entity pk { }":             `Illegal identifier "pk"; reserved words cannot be used as identifiers`,
		"entity t {false: string }": `Illegal identifier "false"; reserved words cannot be used as identifiers`,
		"entity = [1, 2]":           `Illegal identifier "entity"; reserved words cannot be used as identifiers`,
		"generate(1, generate)":     `Illegal identifier "generate"; reserved words cannot be used as identifiers`,
		"[let]":                     `Illegal identifier "let"; reserved words cannot be used as identifiers`,
		"entity t << generate {}":   `Illegal identifier "generate"; reserved words cannot be used as identifiers`,
		"generate << {}":            `Illegal identifier "generate"; reserved words cannot be used as identifiers`,
	}

	for spec, expectedErrMessage := range specs {
		_, err := runParser(spec)
		ExpectsError(t, expectedErrMessage, removeLocationInfo(err))

	}
}
