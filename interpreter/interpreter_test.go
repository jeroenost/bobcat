package interpreter

import "testing"
import "time"
import "fmt"
import "github.com/ThoughtWorksStudios/datagen/dsl"
import "github.com/ThoughtWorksStudios/datagen/generator"

var validFields = []dsl.Node{
	dsl.Node{Kind: "field", Name: "name", Value: "string", Args: stringArgs(10)},
	dsl.Node{Kind: "field", Name: "age", Value: "integer", Args: intArgs(1, 10)},
	dsl.Node{Kind: "field", Name: "thing", Value: "decimal", Args: floatArgs(1, 10)},
	dsl.Node{Kind: "field", Name: "bod", Value: "date", Args: timeArgs("2015-01-01", "2017-01-01")},
	dsl.Node{Kind: "field", Name: "last_name", Value: "dict", Args: dictArgs("last_name")},
}

func TestDefaultArgument(t *testing.T) {
	timeMin, _ := time.Parse("2006-01-02", "1945-01-01")
	timeMax, _ := time.Parse("2006-01-02", "2017-01-01")
	defaults := map[string]interface{}{
		"string":  5,
		"integer": [2]int{1, 10},
		"decimal": [2]float64{1, 10},
		"date":    [2]time.Time{timeMin, timeMax},
		"dict":    "silly_name",
	}

	for kind, expected_value := range defaults {
		actual := defaultArgumentFor(kind)
		if actual != expected_value {
			t.Error(fmt.Sprintf("default value for argument type '%s' was expected to be %v but was %v", kind, expected_value, actual))
		}
	}
}

func TestTranslateFieldsForEntity(t *testing.T) {
	testEntity := generator.NewGenerator("person")
	translateFieldsForEntity(testEntity, validFields)
	for _, field := range validFields {
		if testEntity.GetField == nil {
			t.Errorf("Expected entity to have field %s, but it did not", field.Name)
		}
	}
}

func TestConfiguringFieldForEntity(t *testing.T) {
	testEntity := generator.NewGenerator("person")
	for _, field := range validFields {
		configureFieldOn(testEntity, field)
		if testEntity.GetField(field.Name) == nil {
			t.Errorf("Expected entity to have field %s, but it did not", field.Name)
		}
	}

	if testEntity.GetField("wubba lubba dub dub") != nil {
		t.Error("should not get field for non existent field")
	}
}

func TestValInt(t *testing.T) {
	expected := 666
	actual := valInt(stringArg(666))
	assertExpectedEqsActual(t, expected, actual)
}

func TestValStr(t *testing.T) {
	expected := "blah"
	actual := valStr(dictArg("blah"))
	assertExpectedEqsActual(t, expected, actual)
}

func TestValFloat(t *testing.T) {
	expected := 4.2
	actual := valFloat(floatArg(4.2))
	assertExpectedEqsActual(t, expected, actual)
}

func TestValTime(t *testing.T) {
	expected, _ := time.Parse("2006-01-02", "1945-01-01")
	actual := valTime(timeArg(expected))
	assertExpectedEqsActual(t, expected, actual)
}

func dictArgs(value string) []dsl.Node {
	return []dsl.Node{dictArg(value)}
}

func timeArgs(min, max string) []dsl.Node {
	minTime, _ := time.Parse("20016-01-02", min)
	maxTime, _ := time.Parse("20016-01-02", max)
	return []dsl.Node{timeArg(minTime), timeArg(maxTime)}
}

func stringArgs(value int64) []dsl.Node {
	return []dsl.Node{stringArg(value)}
}

func intArgs(min, max int64) []dsl.Node {
	return []dsl.Node{stringArg(min), stringArg(max)}
}

func floatArgs(min, max float64) []dsl.Node {
	return []dsl.Node{floatArg(min), floatArg(max)}
}

func stringArg(value int64) dsl.Node {
	return dsl.Node{Kind: "literal-int", Value: value}
}

func intArg(value int64) dsl.Node {
	return dsl.Node{Kind: "literal-int", Value: value}
}

func dictArg(value string) dsl.Node {
	return dsl.Node{Kind: "dict", Value: value}
}

func floatArg(value float64) dsl.Node {
	return dsl.Node{Kind: "decimal", Value: value}
}

func timeArg(value time.Time) dsl.Node {
	return dsl.Node{Kind: "date", Value: value}
}

func assertExpectedEqsActual(t *testing.T, expected, actual interface{}) {
	if expected != actual {
		t.Errorf("expected %v, but was %v", expected, actual)
	}
}