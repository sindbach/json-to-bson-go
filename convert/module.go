package convert

import (
	"bytes"
	"fmt"
	"strings"

	jen "github.com/dave/jennifer/jen"
	"github.com/sindbach/json-to-bson-go/options"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"go.mongodb.org/mongo-driver/bson/bsontype"
)

type generatedStruct struct {
	name   string
	fields []jen.Code
}

// Convert is the main entry for the module
func Convert(jsonStr []byte, opts *options.Options) (string, error) {
	if opts == nil {
		opts = options.NewOptions()
	}

	// Set canonical to false, as the only difference for parsing is that canonical extJSON rejects
	// more formats
	ejvr, err := bsonrw.NewExtJSONValueReader(bytes.NewReader(jsonStr), false)
	if err != nil {
		return "", err
	}

	fields, err := getStructFields(ejvr, opts, opts.StructName())
	if err != nil {
		return "", err
	}

	output := jen.NewFile("main")
	output.ImportName("go.mongodb.org/mongo-driver/bson/primitive", "primitive")
	for idx, gs := range fields {
		if idx != 0 {
			output.Line()
		}
		output.Type().Id(gs.name).Struct(gs.fields...)
	}
	return output.GoString(), nil
}

func getStructFields(ejvr bsonrw.ValueReader, opts *options.Options, structName string) ([]generatedStruct, error) {
	var allStructs []generatedStruct
	var topLevelFields []jen.Code

	if ejvr.Type() != bsontype.EmbeddedDocument {
		return nil, fmt.Errorf("expected document type, got %s", ejvr.Type())
	}

	docReader, err := ejvr.ReadDocument()
	if err != nil {
		return nil, err
	}
	key, ejvr, err := docReader.ReadElement()
	if err != nil {
		return nil, err
	}
	for err == nil {
		elemKey := strings.Title(key)
		elem := jen.Id(elemKey)
		structTags := []string{key}

		switch ejvr.Type() {
		case bsontype.Array:
			nestedField, err := getArrayStruct(ejvr, opts, elemKey)
			if err != nil {
				return nil, fmt.Errorf("error processing array for key %q: %w", key, err)
			}
			structTags = append(structTags, "omitempty")

			if nestedField != nil {
				elem.Add(jen.Index().Add(nestedField))
			} else {
				elem.Add(jen.Index().Interface())
			}
		case bsontype.EmbeddedDocument:

			nestedFields, err := getStructFields(ejvr, opts, elemKey)
			if err != nil {
				return nil, fmt.Errorf("error processing nested document for key %q: %w", key, err)
			}
			allStructs = append(allStructs, nestedFields...)
			elem.Add(jen.Id(elemKey))
		default:
			fieldType, addTags, err := getField(ejvr, opts)
			if err != nil {
				return nil, err
			}
			elem.Add(fieldType)
			structTags = append(structTags, addTags...)
			err = ejvr.Skip()
			if err != nil {
				return nil, err
			}
		}

		tagsString := strings.Join(structTags, ",")
		elem.Tag(map[string]string{"bson": tagsString})

		topLevelFields = append(topLevelFields, elem)
		key, ejvr, err = docReader.ReadElement()
	}
	if err != nil && err != bsonrw.ErrEOD {
		return nil, err
	}

	topLevelStruct := generatedStruct{
		name:   structName,
		fields: topLevelFields,
	}
	allStructs = append([]generatedStruct{topLevelStruct}, allStructs...)
	return allStructs, nil
}

func getField(ejvr bsonrw.ValueReader, opts *options.Options) (*jen.Statement, []string, error) {
	structTags := []string{}

	var retVal *jen.Statement
	switch ejvr.Type() {
	case bsontype.Double:
		retVal = jen.Float64()
	case bsontype.String:
		retVal = jen.String()
	case bsontype.Boolean:
		retVal = jen.Bool()
	case bsontype.Int32:
		if !opts.MinimizeIntegerSize() {
			retVal = jen.Float64()
			break
		}
		retVal = jen.Int32()
		if opts.TruncateIntegers() {
			structTags = append(structTags, "truncate")
		}
	case bsontype.Int64:
		if !opts.MinimizeIntegerSize() {
			retVal = jen.Float64()
			break
		}
		retVal = jen.Int64()
		if opts.TruncateIntegers() {
			structTags = append(structTags, "truncate")
		}
	case bsontype.Binary:
		retVal = jen.Qual("go.mongodb.org/mongo-driver/bson/primitive", "Binary")
	case bsontype.Undefined:
		retVal = jen.Qual("go.mongodb.org/mongo-driver/bson/primitive", "Undefined")
	case bsontype.ObjectID:
		retVal = jen.Qual("go.mongodb.org/mongo-driver/bson/primitive", "ObjectID")
	case bsontype.DateTime:
		retVal = jen.Qual("go.mongodb.org/mongo-driver/bson/primitive", "DateTime")
	case bsontype.Null:
		retVal = jen.Interface()
		structTags = append(structTags, "omitempty")
	case bsontype.Regex:
		retVal = jen.Qual("go.mongodb.org/mongo-driver/bson/primitive", "Regex")
	case bsontype.DBPointer:
		retVal = jen.Qual("go.mongodb.org/mongo-driver/bson/primitive", "DBPointer")
	case bsontype.JavaScript:
		retVal = jen.Qual("go.mongodb.org/mongo-driver/bson/primitive", "JavaScript")
	case bsontype.Symbol:
		retVal = jen.Qual("go.mongodb.org/mongo-driver/bson/primitive", "Symbol")
	case bsontype.CodeWithScope:
		retVal = jen.Qual("go.mongodb.org/mongo-driver/bson/primitive", "CodeWithScope")
	case bsontype.Timestamp:
		retVal = jen.Qual("go.mongodb.org/mongo-driver/bson/primitive", "Timestamp")
	case bsontype.Decimal128:
		retVal = jen.Qual("go.mongodb.org/mongo-driver/bson/primitive", "Decimal128")
	case bsontype.MinKey:
		retVal = jen.Qual("go.mongodb.org/mongo-driver/bson/primitive", "MinKey")
	case bsontype.MaxKey:
		retVal = jen.Qual("go.mongodb.org/mongo-driver/bson/primitive", "MaxKey")
	// if we got here, we got here from getArrayStruct and i don't care yet
	case bsontype.EmbeddedDocument, bsontype.Array:
		retVal = jen.Interface()
		structTags = append(structTags, "omitempty")
	default:
		return nil, structTags, fmt.Errorf("Unknown type: %s", ejvr.Type())
	}

	return retVal, structTags, nil
}

func getArrayStruct(ejvr bsonrw.ValueReader, opts *options.Options, name string) (*jen.Statement, error) {
	if ejvr.Type() != bsontype.Array {
		return nil, fmt.Errorf("expected document type, got %s", ejvr.Type())
	}

	arrayReader, err := ejvr.ReadArray()
	if err != nil {
		return nil, err
	}

	var retVal *jen.Statement
	stillChecking := true

	for {
		ejvr, err = arrayReader.ReadValue()
		if err != nil {
			break
		}
		switch ejvr.Type() {
		case bsontype.Array, bsontype.EmbeddedDocument:
			stillChecking = false
			retVal = nil
		default:
			if stillChecking {
				fieldType, _, err := getField(ejvr, opts)
				if err != nil {
					return nil, err
				}
				if retVal == nil {
					retVal = fieldType
				} else if retVal.GoString() != fieldType.GoString() {
					stillChecking = false
					retVal = nil
				}
			}
		}
		err = ejvr.Skip()
		if err != nil {
			return nil, err
		}
	}
	if err != nil && err != bsonrw.ErrEOA {
		return nil, err
	}

	return retVal, nil
}
