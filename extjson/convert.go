package extjson

import (
	"bytes"
	"fmt"
	"strings"

	jen "github.com/dave/jennifer/jen"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"github.com/sindbach/json-to-bson-go/options"
)

func Convert(jsonStr []byte, canonical bool, opts *options.Options) (string, error) {
	if opts == nil {
		opts = options.NewOptions()
	}

	ejvr, err := bsonrw.NewExtJSONValueReader(bytes.NewReader(jsonStr), canonical)
	if err != nil {
		return "", err
	}

	fields, err := getStructFields(ejvr, opts)
	if err != nil {
		return "", err
	}

	output := jen.NewFile("main")
	output.Type().Id(opts.StructName()).Struct(fields...)
	return output.GoString(), nil
}

func getStructFields(ejvr bsonrw.ValueReader, opts *options.Options) ([]jen.Code, error) {
	if ejvr.Type() != bsontype.EmbeddedDocument {
		return nil, fmt.Errorf("expected document type, got %s", ejvr.Type())
	}

	docReader, err := ejvr.ReadDocument()
	if err != nil {
		return nil, err
	}

	var fields []jen.Code
	key, ejvr, err := docReader.ReadElement()
	for err == nil {
		elem := jen.Id(strings.Title(key))
		structTags := []string{key}
		var nestedDoc bool

		switch ejvr.Type() {
		case bsontype.Double:
			elem.Add(jen.Float64())
		case bsontype.String:
			elem.Add(jen.String())
		case bsontype.Boolean:
			elem.Add(jen.Bool())
		case bsontype.Int32:
			if !opts.MinimizeIntegerSize() {
				elem.Add(jen.Float64())
				break
			}
			elem.Add(jen.Int32())
			if opts.TruncateIntegers() {
				structTags = append(structTags, "truncate")
			}
		case bsontype.Int64:
			if !opts.MinimizeIntegerSize() {
				elem.Add(jen.Float64())
				break
			}
			elem.Add(jen.Int64())
			if opts.TruncateIntegers() {
				structTags = append(structTags, "truncate")
			}
		case bsontype.Binary:
			elem.Add(jen.Qual("primitive", "Binary"))
		case bsontype.Undefined:
			elem.Add(jen.Qual("primitive", "Undefined"))
		case bsontype.ObjectID:
			elem.Add(jen.Qual("primitive", "ObjectID"))
		case bsontype.DateTime:
			elem.Add(jen.Qual("primitive", "DateTime"))
		case bsontype.Null:
			elem.Add(jen.Interface())
			structTags = append(structTags, "omitempty")
		case bsontype.Regex:
			elem.Add(jen.Qual("primitive", "Regex"))
		case bsontype.DBPointer:
			elem.Add(jen.Qual("primitive", "DBPointer"))
		case bsontype.JavaScript:
			elem.Add(jen.Qual("primitive", "JavaScript"))
		case bsontype.Symbol:
			elem.Add(jen.Qual("primitive", "Symbol"))
		case bsontype.CodeWithScope:
			elem.Add(jen.Qual("primitive", "CodeWithScope"))
		case bsontype.Timestamp:
			elem.Add(jen.Qual("primitive", "Timestamp"))
		case bsontype.Decimal128:
			elem.Add(jen.Qual("primitive", "Decimal128"))
		case bsontype.MinKey:
			elem.Add(jen.Qual("primitive", "MinKey"))
		case bsontype.MaxKey:
			elem.Add(jen.Qual("primitive", "MaxKey"))
		case bsontype.Array:
			elem.Add(jen.Index().Interface())
			structTags = append(structTags, "omitempty")
		case bsontype.EmbeddedDocument:
			nestedFields, err := getStructFields(ejvr, opts)
			if err != nil {
				return nil, fmt.Errorf("error processing nested document for key %q: %w", key, err)
			}

			elem.Add(jen.Struct(nestedFields...))
			nestedDoc = true
		default:
			return nil, fmt.Errorf("Unknown type: %s", ejvr.Type())
		}

		tagsString := strings.Join(structTags, ",")
		elem.Tag(map[string]string{"bson": tagsString})
		fields = append(fields, elem)
		if !nestedDoc {
			err = ejvr.Skip()
			if err != nil {
				return nil, err
			}
		}
		key, ejvr, err = docReader.ReadElement()
	}

	return fields, nil
}
