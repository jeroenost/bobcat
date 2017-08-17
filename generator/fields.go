package generator

import (
	. "github.com/ThoughtWorksStudios/bobcat/common"
	"github.com/ThoughtWorksStudios/bobcat/dictionary"
	"github.com/rs/xid"
	"math"
	"math/rand"
	"time"
)

var src = rand.NewSource(time.Now().UnixNano())

type Field struct {
	fieldType FieldType
	count     *CountRange
}

func (f *Field) Type() string {
	return f.fieldType.Type()
}

func (f *Field) GenerateValue(parentId string) interface{} {
	if !f.count.Multiple() {
		return f.fieldType.One(parentId)
	} else {
		count := f.count.Count()
		values := make([]interface{}, count)

		for i := int64(0); i < count; i++ {
			values[i] = f.fieldType.One(parentId)
		}

		return values
	}
}

type FieldSet map[string]*Field

type FieldType interface {
	Type() string
	One(parentId string) interface{}
}

func NewField(fieldType FieldType, count *CountRange) *Field {
	return &Field{fieldType: fieldType, count: count}
}

type ReferenceType struct {
	referred  *Generator
	fieldName string
}

func (field *ReferenceType) Type() string {
	return "reference"
}

func (field *ReferenceType) One(parentId string) interface{} {
	ref := field.referred.fields[field.fieldName].fieldType
	return ref.One(parentId)
}

type EntityType struct {
	entityGenerator *Generator
}

func (field *EntityType) Type() string {
	return "entity"
}

func (field *EntityType) One(parentId string) interface{} {
	return field.entityGenerator.One(parentId)
}

type BoolType struct {
}

func (field *BoolType) Type() string {
	return "boolean"
}

func (field *BoolType) One(parentId string) interface{} {
	return 49 < rand.Intn(100)
}

type MongoIDType struct {
}

func (field *MongoIDType) Type() string {
	return "mongoid"
}

func (field *MongoIDType) One(parentId string) interface{} {
	return xid.New().String()
}

type LiteralType struct {
	value interface{}
}

func (field *LiteralType) Type() string {
	return "literal"
}

func (field *LiteralType) One(parentId string) interface{} {
	return field.value
}

type StringType struct {
	length int64
}

func (field *StringType) Type() string {
	return "string"
}

const ALLOWED_CHARACTERS = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@"

var LETTER_INDEX_BITS uint = uint(math.Ceil(math.Log2(float64(len(ALLOWED_CHARACTERS))))) // number of bits to represent ALLOWED_CHARACTERS
var LETTER_BIT_MASK int64 = 1<<LETTER_INDEX_BITS - 1                                      // All 1-bits, as many as LETTER_INDEX_BITS
var LETTERS_PER_INT63 uint = 63 / LETTER_INDEX_BITS                                       // # of letter indices fitting in 63 bits as generated by src.Int63

func (field *StringType) One(parentId string) interface{} {
	n := field.length
	b := make([]byte, n)

	for i, cache, remain := n-1, src.Int63(), LETTERS_PER_INT63; i >= int64(0); {
		if remain == 0 {
			cache, remain = src.Int63(), LETTERS_PER_INT63
		}
		if idx := int(cache & LETTER_BIT_MASK); idx < len(ALLOWED_CHARACTERS) {
			b[i] = ALLOWED_CHARACTERS[idx]
			i--
		}
		cache >>= LETTER_INDEX_BITS
		remain--
	}

	return string(b)
}

type IntegerType struct {
	min int64
	max int64
}

func (field *IntegerType) Type() string {
	return "integer"
}

func (field *IntegerType) One(parentId string) interface{} {
	return field.min + rand.Int63n(field.max-field.min+1)
}

type FloatType struct {
	min float64
	max float64
}

func (field *FloatType) Type() string {
	return "float"
}

func (field *FloatType) One(parentId string) interface{} {
	return rand.Float64()*(field.max-field.min) + field.min
}

type DateType struct {
	min time.Time
	max time.Time
}

func (field *DateType) Type() string {
	return "date"
}

func (field *DateType) ValidBounds() bool {
	return field.min.Before(field.max)
}

func (field *DateType) One(parentId string) interface{} {
	min, max := field.min.Unix(), field.max.Unix()
	delta := max - min
	sec := rand.Int63n(delta) + min

	return time.Unix(sec, 0)
}

type DictType struct {
	category string
}

var CustomDictPath = ""

func (field *DictType) Type() string {
	return "dict"
}

func (field *DictType) One(parentId string) interface{} {
	dictionary.SetCustomDataLocation(CustomDictPath)
	return dictionary.ValueFromDictionary(field.category)
}

type EnumType struct {
	category string
	values   []interface{}
}

func (field *EnumType) Type() string {
	return "enum"
}

func (field *EnumType) One(parentId string) interface{} {
	return field.values[rand.Int63n(int64(len(field.values)))]
}
