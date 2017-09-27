package generator

import (
	. "github.com/ThoughtWorksStudios/bobcat/common"
	. "github.com/ThoughtWorksStudios/bobcat/test_helpers"
	"testing"
	"time"
)

func TestPercentageDistributionOneInteger(t *testing.T) {
	weights := []float64{50.0, 50.0}
	intervalOne := &IntegerType{min: 1, max: 10}
	intervalTwo := &IntegerType{min: 20, max: 30}
	domain := Domain{intervals: []FieldType{intervalOne, intervalTwo}}
	dist := &PercentageDistribution{weights: weights, bins: make([]int64, len(weights))}

	count := 10

	resultIntervalOne := []interface{}{}
	resultIntervalTwo := []interface{}{}

	for i := 0; i < count; i++ {
		v, err := dist.One(domain, nil, nil, nil)
		AssertNil(t, err, "Should not receive error")

		value := v.(int64)

		if value >= intervalOne.min && value <= intervalOne.max {
			resultIntervalOne = append(resultIntervalOne, v)
		} else if value >= intervalTwo.min && value <= intervalTwo.max {
			resultIntervalTwo = append(resultIntervalTwo, v)
		} else {
			t.Errorf("Should not have generated a value outside of the domain!")
		}
	}

	AssertEqual(t, len(resultIntervalOne), 5)
	AssertEqual(t, len(resultIntervalTwo), 5)
}

func TestPercentageDistributionOneLiteralField(t *testing.T) {
	weights := []float64{50.0, 50.0}
	intervalOne := &LiteralType{value: "blah"}
	intervalTwo := &LiteralType{value: "eek"}
	domain := Domain{intervals: []FieldType{intervalOne, intervalTwo}}
	dist := &PercentageDistribution{weights: weights, bins: make([]int64, len(weights))}

	count := 10

	resultIntervalOne := []interface{}{}
	resultIntervalTwo := []interface{}{}

	for i := 0; i < count; i++ {
		v, err := dist.One(domain, nil, nil, nil)
		AssertNil(t, err, "Should not receive error")

		value := v.(string)

		if value == "blah" {
			resultIntervalOne = append(resultIntervalOne, v)
		} else if value == "eek" {
			resultIntervalTwo = append(resultIntervalTwo, v)
		} else {
			t.Errorf("Should not have generated a value outside of the domain!")
		}
	}

	AssertEqual(t, len(resultIntervalOne), 5)
	AssertEqual(t, len(resultIntervalTwo), 5)
}

func TestWeightDistributionOneEnum(t *testing.T) {
	weights := []float64{60.0, 40.0}
	intervalOne := &EnumType{size: 2, values: []interface{}{"one", "two"}}
	intervalTwo := &EnumType{size: 2, values: []interface{}{"three", "four"}}
	domain := Domain{intervals: []FieldType{intervalOne, intervalTwo}}
	dist := &PercentageDistribution{weights: weights, bins: make([]int64, len(weights))}

	count := 10

	resultIntervalOne := []interface{}{}
	resultIntervalTwo := []interface{}{}

	for i := 0; i < count; i++ {
		v, err := dist.One(domain, nil, nil, nil)
		AssertNil(t, err, "Should not receive error")
		value := v.(string)

		if value == "one" || value == "two" {
			resultIntervalOne = append(resultIntervalOne, v)
		} else if value == "three" || value == "four" {
			resultIntervalTwo = append(resultIntervalTwo, v)
		} else {
			t.Errorf("Should not have generated a value outside of the domain!")
		}
	}

	AssertEqual(t, len(resultIntervalOne), 6)
	AssertEqual(t, len(resultIntervalTwo), 4)
}

func TestPercentageDistributionOneDate(t *testing.T) {
	weights := []float64{50.0, 50.0}
	timeMin, _ := time.Parse("2006-01-02", "1945-01-01")
	timeMax, _ := time.Parse("2006-01-02", "1945-01-02")
	timeMax2, _ := time.Parse("2006-01-02", "1950-01-02")
	intervalOne := &DateType{min: timeMin, max: timeMax}
	intervalTwo := &DateType{min: timeMax, max: timeMax2}
	domain := Domain{intervals: []FieldType{intervalOne, intervalTwo}}
	dist := &PercentageDistribution{weights: weights, bins: make([]int64, len(weights))}

	count := 10

	resultIntervalOne := []interface{}{}
	resultIntervalTwo := []interface{}{}

	for i := 0; i < count; i++ {
		v, err := dist.One(domain, nil, nil, nil)
		AssertNil(t, err, "Should not receive error")

		value := v.(*TimeWithFormat).Time

		if value.After(intervalOne.min) && value.Before(intervalOne.max) {
			resultIntervalOne = append(resultIntervalOne, v)
		} else if value.After(intervalTwo.min) && value.Before(intervalTwo.max) {
			resultIntervalTwo = append(resultIntervalTwo, v)
		} else {
			t.Errorf("Should not have generated a value outside of the domain! %v\n", value)
		}
	}

	AssertEqual(t, len(resultIntervalOne), 5)
	AssertEqual(t, len(resultIntervalTwo), 5)
}

func TestWeightDistributionOne(t *testing.T) {
	weights := []float64{50.0, 50.0}
	intervalOne := &IntegerType{min: 1, max: 10}
	intervalTwo := &IntegerType{min: 20, max: 30}
	domain := Domain{intervals: []FieldType{intervalOne, intervalTwo}}
	dist := &WeightDistribution{weights: weights}

	count := 10

	resultIntervalOne := []interface{}{}
	resultIntervalTwo := []interface{}{}

	for i := 0; i < count; i++ {
		v, err := dist.One(domain, nil, nil, nil)
		AssertNil(t, err, "Should not receive error")
		value := v.(int64)

		if value >= intervalOne.min && value <= intervalOne.max {
			resultIntervalOne = append(resultIntervalOne, v)
		} else if value >= intervalTwo.min && value <= intervalTwo.max {
			resultIntervalTwo = append(resultIntervalTwo, v)
		} else {
			t.Errorf("Should not have generated a value outside of the domain!")
		}
	}

	Assert(t, len(resultIntervalOne) > 0, "expected to generate at least one")
	Assert(t, len(resultIntervalTwo) > 0, "expected to generate at least one")
}

func TestNormalCompatibleDomain(t *testing.T) {
	norm := &NormalDistribution{}
	Assert(t, norm.isCompatibleDomain(FLOAT_TYPE), "floats should be a compatible domain for normal distributions")
	Assert(t, !norm.isCompatibleDomain(INT_TYPE), "ints should not be a compatible domain for normal distributions")
}

func TestUniformCompatibleDomain(t *testing.T) {
	uni := &UniformDistribution{}
	Assert(t, uni.isCompatibleDomain(FLOAT_TYPE), "floats should be a compatible domain for uniform distributions")
	Assert(t, uni.isCompatibleDomain(INT_TYPE), "ints should be a compatible domain for uniform distributions")
	Assert(t, !uni.isCompatibleDomain(STRING_TYPE), "strings should not be a compatible domain for uniform distributions")
}

func TestNormalShouldntSupportMultipleIntervals(t *testing.T) {
	norm := &NormalDistribution{}
	Assert(t, !norm.supportsMultipleIntervals(), "normal distributions don't support multiple domains")
}

func TestUniformShouldntSupportMultipleIntervals(t *testing.T) {
	uni := &UniformDistribution{}
	Assert(t, !uni.supportsMultipleIntervals(), "uniform distributions don't support multiple domains")
}

func TestPercentageShouldSupportMultipleIntervals(t *testing.T) {
	w := &PercentageDistribution{}
	Assert(t, w.supportsMultipleIntervals(), "percent distributions should support multiple domains")
}

func TestWeightedShouldSupportMultipleIntervals(t *testing.T) {
	w := &WeightDistribution{}
	Assert(t, w.supportsMultipleIntervals(), "weight distributions should support multiple domains")
}

func TestWeightedType(t *testing.T) {
	w := &WeightDistribution{}
	AssertEqual(t, WEIGHT_DIST, w.Type())
}

func TestPercentType(t *testing.T) {
	w := &PercentageDistribution{}
	AssertEqual(t, PERCENT_DIST, w.Type())
}

func TestNormalType(t *testing.T) {
	w := &NormalDistribution{}
	AssertEqual(t, NORMAL_DIST, w.Type())
}

func TestUniformType(t *testing.T) {
	w := &UniformDistribution{}
	AssertEqual(t, UNIFORM_DIST, w.Type())
}
