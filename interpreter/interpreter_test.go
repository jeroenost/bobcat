package interpreter

import (
	. "github.com/ThoughtWorksStudios/bobcat/common"
	"github.com/ThoughtWorksStudios/bobcat/dsl"
	. "github.com/ThoughtWorksStudios/bobcat/emitter"
	"github.com/ThoughtWorksStudios/bobcat/generator"
	. "github.com/ThoughtWorksStudios/bobcat/test_helpers"
	"testing"
)

func AssertShouldHaveField(t *testing.T, entity *generator.Generator, fieldName string, scope *Scope) {
	emitter := NewDummyEmitter()
	result, err := entity.One(nil, emitter, scope)

	AssertNil(t, err, "Should not receive error")
	AssertNotNil(t, result[fieldName], "Expected entity to have field %s, but it did not", fieldName)
}

func AssertFieldYieldsValue(t *testing.T, entity *generator.Generator, field *Node, scope *Scope) {
	emitter := NewDummyEmitter()
	result, err := entity.One(nil, emitter, scope)

	AssertNil(t, err, "Should not receive error")
	AssertEqual(t, field.ValNode().Value, result[field.Name])
}

var validFields = NodeSet{
	Field("name", CallNode(nil, Builtin(STRING_TYPE), IntArgs(10))),
	Field("age", CallNode(nil, Builtin(INT_TYPE), IntArgs(1, 10))),
	Field("weight", CallNode(nil, Builtin(FLOAT_TYPE), FloatArgs(1.0, 200.0))),
	Field("dob", CallNode(nil, Builtin(DATE_TYPE), DateArgs("2015-01-01", "2017-01-01"))),
	Field("last_name", CallNode(nil, Builtin(DICT_TYPE), StringArgs("last_name"))),
	Field("status", CallNode(nil, Builtin(ENUM_TYPE), NodeSet{StringCollection("enabled", "disabled")})),
	Field("status", CallNode(nil, Builtin(SERIAL_TYPE), NodeSet{})),
	Field("catch_phrase", StringVal("Grass.... Tastes bad")),
}

var nestedFields = NodeSet{
	Field("name", CallNode(nil, Builtin(STRING_TYPE), IntArgs(10))),
	Field("pet", Id("Goat")),
	Field("friend", Entity("Horse", validFields)),
}

var overridenFields = NodeSet{
	Field("catch_phrase", StringVal("Grass.... Tastes good")),
}

func interp() *Interpreter {
	emitter := NewDummyEmitter()
	return New(emitter, false)
}

func TestScopingResolvesOtherEntities(t *testing.T) {
	scope := NewRootScope()
	i := interp()
	node := Root(Entity("person", NodeSet{
		Field("pet", Entity("kitteh", overridenFields)),
		Field("pets_can_have_pets_too", Entity("lolcat", NodeSet{
			Field("cheezburgrz", StringVal("can has")),
			Field("protoype", Id("kitteh")),
		})),
	}))
	_, err := i.Visit(node, scope, false)
	AssertNil(t, err, "`lolcat` should be able to resolve `kitteh` because it lives within the scope hierarchy. error was %v", err)

	// using same root scope to simulate previously defined symbols
	_, err = i.Visit(Root(Generation(2, Id("person"))), scope, false)
	AssertNil(t, err, "Should be able to resolve `person` because it is defined in the root scope. error was %v", err)

	// using same root scope to simulate previously defined symbols; here, `kitteh` was defined in a child scope of `person`,
	// but not at the root scope, so we should not be able to resolve it.
	_, err = i.Visit(Root(Generation(1, Id("kitteh"))), scope, false)
	ExpectsError(t, "Cannot resolve symbol \"kitteh\"", err)
}

func TestValidVisit(t *testing.T) {
	node := Root(Entity("person", validFields), Generation(2, Id("person")))
	i := interp()
	scope := NewRootScope()
	_, err := i.Visit(node, scope, false)
	if err != nil {
		t.Errorf("There was a problem generating entities: %v", err)
	}

	for _, entry := range scope.Symbols {
		entity := entry.(*generator.Generator)
		for _, field := range validFields {
			AssertShouldHaveField(t, entity, field.Name, scope)
		}
	}
}

func TestValidVisitWithNesting(t *testing.T) {
	node := Root(Entity("Goat", validFields), Entity("person", nestedFields),
		Generation(2, Id("person")))
	i := interp()

	scope := NewRootScope()
	_, err := i.Visit(node, scope, false)
	if err != nil {
		t.Errorf("There was a problem generating entities: %v", err)
	}

	person, _ := i.ResolveEntityFromNode(Id("person"), scope)
	for _, field := range nestedFields {
		AssertShouldHaveField(t, person, field.Name, scope)
	}
}

func TestValidVisitWithOverrides(t *testing.T) {
	node := Root(
		Entity("person", validFields),
		Generation(
			2,
			EntityExtension("lazyPerson", "person", overridenFields),
		),
	)

	i := interp()
	scope := NewRootScope()

	if _, err := i.Visit(node, scope, false); err != nil {
		t.Errorf("There was a problem generating entities: %v", err)
	}

	AssertEqual(t, 2, len(scope.Symbols), "Should have 2 entities defined")

	for _, key := range []string{"person", "lazyPerson"} {
		_, isPresent := scope.Symbols[key]
		// don't try to use AssertNotNil here; it won't work because it is unable to detect
		// whether a nil pointer passed as an interface{} param to AssertNotEqual is nil.
		// see this crazy shit: https://stackoverflow.com/questions/13476349/check-for-nil-and-nil-interface-in-go
		Assert(t, isPresent, "`%v` should be defined in scope", key)

		if isPresent {
			entity, isGeneratorType := scope.Symbols[key].(*generator.Generator)
			Assert(t, isGeneratorType, "`key` should be defined")

			if key != "person" {
				for _, field := range overridenFields {
					AssertFieldYieldsValue(t, entity, field, scope)
				}
			}
		}
	}
}

func TestPrimaryKey(t *testing.T) {
	for _, script := range []string{
		`
    pk("primary_key", $incr)
    entity {}
    `, // top-level statement
		`entity { pk("primary_key", $incr) }`, // within entity expression
	} {
		scope := NewRootScope()
		ast, err := dsl.Parse("filename", []byte(script))
		AssertNil(t, err, "Should not receive error while parsing")

		i := interp()
		actual, err := i.Visit(ast.(*Node), scope, false)
		AssertNil(t, err, "Should not receive error while interpreting")

		gen, ok := actual.(*generator.Generator)
		Assert(t, ok, "Should have returned a generator, but was %T %v", actual, actual)

		AssertEqual(t, "primary_key", gen.PrimaryKeyName())
		Assert(t, gen.HasField("primary_key"), "Should have primary key field")

		pk_field := gen.GetField("primary_key")
		val, err := pk_field.GenerateValue(nil, nil, nil)
		AssertNil(t, err, "Should not receive error while generating pk value")
		AssertEqual(t, uint64(0), val)
	}
}

func TestDeferredEvaluation(t *testing.T) {
	scope := NewRootScope()
	scope.SetSymbol("foo", int64(10))

	ast, err := dsl.Parse("testScript", []byte(`1 + 2 + 4 * foo * (foo + 18) - foo`))
	AssertNil(t, err, "Should not receive error while parsing")

	i := interp()
	actual, err := i.Visit(ast.(*Node), scope, true)
	AssertNil(t, err, "Should not receive error while interpreting")

	result, ok := actual.(DeferredResolver)
	Assert(t, ok, "Should return a DeferredResolver")

	val, err := result(scope)
	AssertNil(t, err, "Should not receive error while evaluating resolver")

	AssertEqual(t, int64(1113), val)
}

type EvalSpec map[string]interface{}

func TestBinaryExpressionComposition(t *testing.T) {
	i := interp()
	scope := NewRootScope()

	for expr, expected := range (EvalSpec{
		"1 + 2 * 3":                        int64(7),
		"1 * 2 + 3":                        int64(5),
		"(1 + 2) * 3":                      int64(9),
		"5 * 2":                            int64(10),
		"5 / 2":                            float64(2.5),
		"5.0 / 2":                          float64(2.5),
		"5 / 2.0":                          float64(2.5),
		"(\"hi \" + \"thar\" + 5) + false": "hi thar5false",
		"3 * \"hi\"":                       "hihihi",
		"\"hi\" * 3":                       "hihihi",
		"5 * 3.0":                          float64(15),
		"3.0 * 5":                          float64(15),
		"true + \" that\"":                 "true that",
		"1 + 2 + 4 * 10 * (10 + 18) - 10":  int64(1113),
		"(-2 * (6 - 7) / 2) * 88 / 4":      float64(22),
	}) {
		ast, err := dsl.Parse("testScript", []byte(expr))
		AssertNil(t, err, "Should not receive error while parsing %q", expr)

		actual, err := i.Visit(ast.(*Node), scope, false)
		AssertNil(t, err, "Should not receive error while interpreting %q", expr)

		AssertEqual(t, expected, actual, "Incorrect result for %q", expr)
	}
}

func TestBinaryExpressionAsEntityField(t *testing.T) {
	i := interp()
	scope := NewRootScope()
	expr := "entity foo { field: 1 + 1 }"

	ast, err := dsl.Parse("testScript", []byte(expr))
	AssertNil(t, err, "Should not receive error while parsing %q", expr)

	actual, err := i.Visit(ast.(*Node), scope, false)
	AssertNil(t, err, "Should not receive error while interpreting %q", expr)

	entity := actual.(*generator.Generator)
	expectedFieldName := "field"

	Assert(t, entity.HasField(expectedFieldName), "Field %q does not exist", expectedFieldName)
}

func TestComplexExpressionFieldEvaluation(t *testing.T) {
	i := interp()
	scope := NewRootScope()
	expr := `
  let standard_tip = 0.15

  entity RestaurantBill {
    price: $float(1.0, 10.0),
    tax: (let sf_tax = 0.085; lambda perc(amount, rate) { amount * rate })(price, sf_tax),

    # validate we can refer to perc() field type is a binary expression
    total: price + tax + perc(price, standard_tip)
  }`

	ast, err := dsl.Parse("testScript", []byte(expr))
	AssertNil(t, err, "Should not receive error while parsing %q", expr)

	actual, err := i.Visit(ast.(*Node), scope, false)
	AssertNil(t, err, "Should not receive error while interpreting %q", expr)

	entity, err := actual.(*generator.Generator).One("", NewDummyEmitter(), scope)
	AssertNil(t, err, "Should not receive error")

	price := entity["price"].(float64)
	Assert(t, price >= 1.0 && price <= 10.0, "Should generate price within bounds")

	tax := entity["tax"].(float64)
	AssertEqual(t, RoundFloat(price*0.085, 0.01), RoundFloat(tax, 0.01), "Should calculate tax properly")

	total := entity["total"].(float64)
	AssertEqual(t, RoundFloat(price+tax+(price*0.15), 0.01), RoundFloat(total, 0.01), "Should calculate total properly")
}

func TestCallableExpressionFieldCanReferenceDeclaredLambdaInPriorField(t *testing.T) {
	i := interp()
	scope := NewRootScope()
	expr := `
  entity Foo {
    price: $float(1.0, 30.0),
    tax: (lambda perc(amount, rate) {amount * rate})(price, 0.085),

    # validate we can refer to perc() when field type is a call expression
    # and callee.Kind == "identifier"
    tip: perc(price, 0.15),

    total: price + tax + tip
  }`

	ast, err := dsl.Parse("testScript", []byte(expr))
	AssertNil(t, err, "Should not receive error while parsing %q", expr)

	actual, err := i.Visit(ast.(*Node), scope, false)
	AssertNil(t, err, "Should not receive error while interpreting %q", expr)

	entity, err := actual.(*generator.Generator).One("", NewDummyEmitter(), scope)
	AssertNil(t, err, "Should not receive error")

	price := entity["price"].(float64)
	Assert(t, price >= 1.0 && price <= 30.0, "Should generate price within bounds")

	tax := entity["tax"].(float64)
	AssertEqual(t, RoundFloat(price*0.085, 0.01), RoundFloat(tax, 0.01), "Should calculate tax properly")

	tip := entity["tip"].(float64)
	AssertEqual(t, RoundFloat(price*0.15, 0.01), RoundFloat(tip, 0.01), "Should calculate tip properly")

	total := entity["total"].(float64)
	AssertEqual(t, RoundFloat(price+tax+tip, 0.01), RoundFloat(total, 0.01), "Should calculate total properly")
}

func TestLambdaExpression(t *testing.T) {
	scope := NewRootScope()

	script := `
  lambda Square(x) {
    x * x
  }

  let foo = 2, bar = 4

  # demonstrate nested call within inlined call with closure
  (lambda () {
    foo = bar * foo
    Square(foo)
  })()
  `
	ast, err := dsl.Parse("testScript", []byte(script))
	AssertNil(t, err, "Should not receive error while parsing")

	i := interp()
	actual, err := i.Visit(ast.(*Node), scope, false)
	AssertNil(t, err, "Should not receive error while interpreting")

	AssertEqual(t, int64(64), actual, "Unexpected result %T %v", actual, actual)
}

func TestLambdaExpressionNoOp(t *testing.T) {
	scope := NewRootScope()

	script := `
  lambda noop(x) {}
  noop(5)
  `
	ast, err := dsl.Parse("testScript", []byte(script))
	AssertNil(t, err, "Should not receive error while parsing")

	i := interp()
	actual, err := i.Visit(ast.(*Node), scope, false)
	AssertNil(t, err, "Should not receive error while interpreting")
	AssertNil(t, actual, "noop() should do nothing and return nil")
}

func TestLambdaExpressionVariableShadowing(t *testing.T) {
	scope := NewRootScope()

	script := `
  let foo = 1

  lambda BoundParamShadows(foo) {
    "bound lambda arg 'foo' is " + foo
  }

  lambda VarDeclShadows(x) {
    let foo = x
    "declared variable 'foo' within lambda is " + foo
  }

  let shadowed = BoundParamShadows(5) + ", " + VarDeclShadows(10)

  shadowed + ", " + "but outer scoped 'foo' is still " + foo
  `
	ast, err := dsl.Parse("testScript", []byte(script))
	AssertNil(t, err, "Should not receive error while parsing")

	i := interp()

	expected := "bound lambda arg 'foo' is 5, declared variable 'foo' within lambda is 10, but outer scoped 'foo' is still 1"
	actual, err := i.Visit(ast.(*Node), scope, false)

	AssertNil(t, err, "Should not receive error while interpreting")
	AssertEqual(t, expected, actual)
}

func TestLambdaExpressionsAllowComments(t *testing.T) {
	scope := NewRootScope()
	script := `
  let baz = 10

  lambda test1() {
    let foo = 1, bar = 2 # multiple declarations with comment

    baz = 0 # don't break


    # this comment shouldn't break anything
    foo + bar # nor should a terminal comment
  }

  lambda test2() {
    # shouldn't break when first token of lambda body is a comment
    test1()
    # sequential expressions should still work too; this one should return 6
    4, 5, 6
  }

  test1() + test2() # comments outside of lambdas should still be ok
  `

	ast, err := dsl.Parse("testScript", []byte(script))
	AssertNil(t, err, "Should not receive error while parsing; comments may be interfering with parsing.")

	i := interp()

	actual, err := i.Visit(ast.(*Node), scope, false)

	AssertNil(t, err, "Should not receive error while interpreting")
	AssertEqual(t, int64(9), actual, "Comments should not affect interpretation of lambdas and calls")
}

func TestLambdaExpressionUsesStaticScoping(t *testing.T) {
	scope := NewRootScope()

	script := `
  let b = "static"

  lambda fn() {
    lambda foo() {
      b + " scoping"
    }

    lambda bar() { # test that foo has static scope
      let b = "dynamic"
      foo()
    }
  }

  lambda baz() { # test that fn has static scope
    let b = "dynamic"
    (fn())()
  }

  baz() # should invoke bar()
`
	ast, err := dsl.Parse("testScript", []byte(script))
	AssertNil(t, err, "Should not receive error while parsing")

	i := interp()

	actual, err := i.Visit(ast.(*Node), scope, false)

	AssertNil(t, err, "Should not receive error while interpreting")
	AssertEqual(t, "static scoping", actual, "Should be using a lexical/static scope for lambda declarations")
}

func TestLambdaExpressionsWithClosuresContinueToWorkAfterFirstInvocation(t *testing.T) {
	script := `
  let foo

  lambda outer() {
    foo = 1

    lambda inner() {
      foo = foo * 2
    }
  }

  let pow2 = outer()

  pow2() # foo => 2
  pow2() # foo => 4; there was a bug that prevented inner() from executing its body when invoked more than once
  pow2() # foo => 8; there was a bug that prevented inner() from executing its body when invoked more than once

  foo
  `
	scope := NewRootScope()
	ast, err := dsl.Parse("testScript", []byte(script))
	AssertNil(t, err, "Should not receive error while parsing")

	i := interp()

	actual, err := i.Visit(ast.(*Node), scope, false)

	AssertNil(t, err, "Should not receive error while interpreting")
	AssertEqual(t, int64(8), actual, "Lambda closures should continue to work after first invocation when symbols are involved")
}

func TestAutomaticGenerateOnEntityValues(t *testing.T) {
	for _, script := range []string{
		`entity Foo { sub: *weight ~ [1 => entity { pk("id", $uid) }] }`,
		`entity Foo { sub: $enum([entity { pk("id", $uid) }]) }`,
		`entity Foo { sub: (lambda mkEntity(pkName) { entity { pk(pkName, $uid) } })("id") }`,
	} {
		scope := NewRootScope()
		emitter := NewTestEmitter()
		ast, err := dsl.Parse("testScript", []byte(script))
		AssertNil(t, err, "Should not receive error while parsing")

		result, err := interp().Visit(ast.(*Node), scope, false)
		AssertNil(t, err, "Should not receive error while interpreting")

		outer, ok := result.(*generator.Generator)
		Assert(t, ok, "Expected a Generator value, but got %T", result)

		_, err = outer.One(nil, emitter, scope)
		AssertNil(t, err, "Should not receive error during generation")

		sub := emitter.Shift()
		entityResult := emitter.Shift()

		AssertEqual(t, sub["id"], entityResult["sub"], "Should have generated the subentity, and the IDs should match")
		AssertEqual(t, sub["$parent"], entityResult["$id"], "Should have generated the subentity, and the IDs should match")

		first := entityResult["sub"]

		// Do this again to ensure we're not just storing a static value, but generating each time
		_, err = outer.One(nil, emitter, scope)
		AssertNil(t, err, "Should not receive error during generation")

		sub = emitter.Shift()
		entityResult = emitter.Shift()

		AssertEqual(t, sub["id"], entityResult["sub"], "Should have generated the subentity, and the IDs should match")
		AssertEqual(t, sub["$parent"], entityResult["$id"], "Should have generated the subentity, and the IDs should match")

		second := entityResult["sub"]

		AssertNotEqual(t, first, second, "Should prove that entity is generated every time, and that we aren't storing a static value")
	}
}

func TestBuiltinLambdaEvaluation(t *testing.T) {
	script := `
	  let fn = $int
	  fn(1,4)
	`
	scope := NewRootScope()
	ast, err := dsl.Parse("testScript", []byte(script))
	AssertNil(t, err, "Should not receive error while parsing")

	i := interp()

	result, err := i.Visit(ast.(*Node), scope, false)

	AssertNil(t, err, "Should not receive error while interpreting")
	actual, ok := result.(int64)

	Assert(t, ok, "Should have returned an integer")
	Assert(t, actual <= int64(4) && actual >= int64(1), "Should return an integer between 1 and 4")
}

func TestValidGenerationNodeIdentifierAsCountArg(t *testing.T) {
	i := interp()
	scope := NewRootScope()
	i.EntityFromNode(Entity("person", validFields), scope, false)
	scope.SetSymbol("count", int64(1))
	node := GenNode(nil, NodeSet{Id("count"), Id("person")})
	_, err := i.GenerateFromNode(node, scope, false)
	AssertNil(t, err, "Should be able to use identifiers as count argument")
}

func TestInvalidGenerationNodeBadCountArg(t *testing.T) {
	i := interp()
	scope := NewRootScope()
	i.EntityFromNode(Entity("person", validFields), scope, false)
	node := Generation(0, Id("person"))
	_, err := i.GenerateFromNode(node, scope, false)
	ExpectsError(t, "Must generate at least 1 person{} entity", err)

	scope.SetSymbol("count", "ten")
	node = GenNode(nil, NodeSet{Id("count"), Id("person")})
	_, err = i.GenerateFromNode(node, scope, false)
	ExpectsError(t, "Expected an integer, but got ten", err)
}

func TestEntityWithUndefinedParent(t *testing.T) {
	ent := Entity("person", validFields)
	unresolvable := Id("nope")
	ent.Related = unresolvable
	_, err := interp().EntityFromNode(ent, NewRootScope(), false)
	ExpectsError(t, `Cannot resolve parent entity "nope" for entity "person"`, err)
}

func TestGenerateEntitiesCannotResolveEntityFromNode(t *testing.T) {
	node := Generation(2, Id("tree"))
	_, err := interp().GenerateFromNode(node, NewRootScope(), false)
	ExpectsError(t, `Cannot resolve symbol "tree"`, err)
}

func TestDisallowNondeclaredEntityAsFieldIdentifier(t *testing.T) {
	i := interp()
	_, e := i.EntityFromNode(Entity("hiccup", nestedFields), NewRootScope(), false)
	ExpectsError(t, `Cannot resolve symbol "Goat"`, e)

}

func TestConfiguringFieldDiesWhenFieldWithoutArgsHasNoDefaults(t *testing.T) {
	ast, err := dsl.Parse("testScript", []byte("generate(1, entity { name: $dict() })"))
	_, err = interp().Visit(ast.(*Node), NewRootScope(), false)

	ExpectsError(t, "Usage: $dict(category_name)", err)
}

func TestConfiguringFieldWithoutArguments(t *testing.T) {
	ast, err := dsl.Parse("testScript", []byte("entity { last_name: $str() }"))
	scope := NewRootScope()
	result, err := interp().Visit(ast.(*Node), scope, false)

	AssertNil(t, err)

	testEntity, ok := result.(*generator.Generator)

	Assert(t, ok, "Should yield an entity")
	AssertShouldHaveField(t, testEntity, "last_name", scope)
}

func TestConfiguringFieldsForEntityErrors(t *testing.T) {
	ast, err := dsl.Parse("testScript", []byte(`generate(1, entity { name: $dict("foo", "bar") })`))
	_, err = interp().Visit(ast.(*Node), NewRootScope(), false)

	ExpectsError(t, "Usage: $dict(category_name)", err)
}

func TestCanResolvePreviousFieldsIfDefined(t *testing.T) {
	scope := NewRootScope()
	i := interp()
	price := 2.0
	node := Root(Entity("cart", NodeSet{
		Field("price", FloatVal(price)),
		Field("price_clone", Id("price")),
	}))

	entity, _ := i.Visit(node, scope, false)
	resolvedEntity, err := entity.(*generator.Generator).One(nil, NewTestEmitter(), scope)

	AssertNil(t, err, "Should not receive error")
	AssertEqual(t, price, resolvedEntity["price_clone"])
}

func TestThrowsErrorIfCannotResolveSymbolInFieldDeclaration(t *testing.T) {
	scope := NewRootScope()
	i := interp()
	node := Root(Entity("cart", NodeSet{
		Field("price_clone", Id("price")),
	}))

	_, err := i.Visit(node, scope, false)

	ExpectsError(t, "Cannot resolve symbol \"price\"", err)
}

func TestConfiguringDistributionWithStaticFields(t *testing.T) {
	i := interp()
	testEntity := generator.NewGenerator("person", nil, false)

	field := Field("age", Distribution(PERCENT_DIST, AssociativeArgumentNode(nil, FloatLiteralNode(nil, 1), StringVal("blah"))))
	scope := NewRootScope()
	AssertNil(t, i.withDistributionField(testEntity, field, scope, false), "Should not receive an error adding a distro field")
	AssertShouldHaveField(t, testEntity, field.Name, scope)
}

func TestConfiguringDistributionWithMixedFieldTypes(t *testing.T) {
	i := interp()
	testEntity := generator.NewGenerator("person", nil, false)

	value1 := StringVal("disabled")
	value2 := CallNode(nil, Builtin(ENUM_TYPE), NodeSet{StringCollection("enabled", "pending")})

	arg1 := AssociativeArgumentNode(nil, IntLiteralNode(nil, 1), value1)
	arg2 := AssociativeArgumentNode(nil, IntLiteralNode(nil, 2), value2)

	field := Field("age", Distribution(WEIGHT_DIST, arg1, arg2))
	scope := NewRootScope()
	AssertNil(t, i.withDistributionField(testEntity, field, scope, false), "Should not receive an error adding a distro field")
	AssertShouldHaveField(t, testEntity, field.Name, scope)
}

func TestConfiguringDistributionWithEntityField(t *testing.T) {
	i, scope := interp(), NewRootScope()
	testEntity := generator.NewGenerator("person", nil, false)
	scope.SetSymbol("Goat", generator.NewGenerator("goat", nil, false))

	arg1 := AssociativeArgumentNode(nil, IntLiteralNode(nil, 1), Entity("Horse", validFields))
	arg2 := AssociativeArgumentNode(nil, IntLiteralNode(nil, 1), Id("Goat"))

	field := Field("friend", Distribution(WEIGHT_DIST, arg1, arg2))
	AssertNil(t, i.withDistributionField(testEntity, field, scope, false), "Should not receive an error adding a distro field")
	AssertShouldHaveField(t, testEntity, field.Name, scope)
}
