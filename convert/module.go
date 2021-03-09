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

// ImportPrimitive is a constant for bson.primitive module
const ImportPrimitive string = "go.mongodb.org/mongo-driver/bson/primitive"

type snippet struct {
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

	snippets, err := processDocument(ejvr, opts, opts.StructName())
	if err != nil {
		return "", err
	}

	output := jen.NewFile("main")
	output.ImportName(ImportPrimitive, "primitive")
	for idx, structSnippet := range snippets {
		if idx != 0 {
			output.Line()
		}
		output.Type().Id(structSnippet.name).Struct(structSnippet.fields...)
	}
	return output.GoString(), nil
}

func processDocument(ejvr bsonrw.ValueReader, opts *options.Options, structName string) ([]snippet, error) {
	var result []snippet
	var topLevelFields []jen.Code

	if ejvr.Type() != bsontype.EmbeddedDocument {
		return nil, fmt.Errorf("Expecting a document type, received %s", ejvr.Type())
	}
	docReader, err := ejvr.ReadDocument()
	if err != nil {
		return nil, err
	}
	key, ejvr, err := docReader.ReadElement()
	if err != nil {
		return nil, err
	}
	// loop through the document
	for err == nil {
		elemKey := strings.Title(key)
		elem := jen.Id(elemKey)
		structTags := []string{key}

		switch ejvr.Type() {
		case bsontype.Array:
			arrayField, snippetResult, err := processArray(ejvr, opts, elemKey)
			if err != nil {
				return nil, fmt.Errorf("error processing array for key %q: %w", key, err)
			}
			structTags = append(structTags, "omitempty")
			fmt.Printf("processArray returns snippet: %s \n", snippetResult)
			if snippetResult.name != "" {
				result = append(result, snippetResult)
			}
			elem.Add(arrayField)
		case bsontype.EmbeddedDocument:
			nestedFields, err := processDocument(ejvr, opts, elemKey)
			if err != nil {
				return nil, fmt.Errorf("error processing nested document for key %q: %w", key, err)
			}
			result = append(result, nestedFields...)
			elem.Add(jen.Id(elemKey))
		default:
			fieldType, addTags, err := processField(ejvr, opts)
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
	topLevelStruct := snippet{
		name:   structName,
		fields: topLevelFields,
	}
	result = append([]snippet{topLevelStruct}, result...)
	return result, nil
}

func processField(ejvr bsonrw.ValueReader, opts *options.Options) (*jen.Statement, []string, error) {
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
		retVal = jen.Qual(ImportPrimitive, "Binary")
	case bsontype.Undefined:
		retVal = jen.Qual(ImportPrimitive, "Undefined")
	case bsontype.ObjectID:
		retVal = jen.Qual(ImportPrimitive, "ObjectID")
	case bsontype.DateTime:
		retVal = jen.Qual(ImportPrimitive, "DateTime")
	case bsontype.Null:
		retVal = jen.Interface()
		structTags = append(structTags, "omitempty")
	case bsontype.Regex:
		retVal = jen.Qual(ImportPrimitive, "Regex")
	case bsontype.DBPointer:
		retVal = jen.Qual(ImportPrimitive, "DBPointer")
	case bsontype.JavaScript:
		retVal = jen.Qual(ImportPrimitive, "JavaScript")
	case bsontype.Symbol:
		retVal = jen.Qual(ImportPrimitive, "Symbol")
	case bsontype.CodeWithScope:
		retVal = jen.Qual(ImportPrimitive, "CodeWithScope")
	case bsontype.Timestamp:
		retVal = jen.Qual(ImportPrimitive, "Timestamp")
	case bsontype.Decimal128:
		retVal = jen.Qual(ImportPrimitive, "Decimal128")
	case bsontype.MinKey:
		retVal = jen.Qual(ImportPrimitive, "MinKey")
	case bsontype.MaxKey:
		retVal = jen.Qual(ImportPrimitive, "MaxKey")
	// if we got here, we got here from getArrayStruct and i don't care yet
	case bsontype.EmbeddedDocument, bsontype.Array:
		retVal = jen.Interface()
		structTags = append(structTags, "omitempty")
	default:
		return nil, structTags, fmt.Errorf("Unknown type: %s", ejvr.Type())
	}

	return retVal, structTags, nil
}

func processArray(ejvr bsonrw.ValueReader, opts *options.Options, name string) (*jen.Statement, snippet, error) {

	/*
		 handles:
		 	a :[1, 2] => int32
			a :[1, "a"] => interface

			a :[{b:1}, {b:2}] =>  []A
			a :[{b:1}, {b:"a"}] => []interface
			a :[{b:[1,2], {b:[1,2]}}] => []A where b is []interface
			a :[{b:{c:1}}, {b:{c:2}}] => []A where b is []interface

		// Always pick the first document/array in array

	*/
	snippetResult := snippet{}
	// Default return value is interface.
	result := jen.Index().Interface()
	if ejvr.Type() != bsontype.Array {
		return result, snippetResult, fmt.Errorf("Expecting an array type, received %s", ejvr.Type())
	}
	arrayReader, err := ejvr.ReadArray()
	if err != nil {
		return result, snippetResult, err
	}

	var retVal *jen.Statement
	stillChecking := true
	nested := false
	for {
		ejvr, err = arrayReader.ReadValue()
		if err != nil {
			fmt.Println(err)
			break
		}
		switch ejvr.Type() {

		// Array of array
		case bsontype.Array:
			stillChecking = false
			retVal = nil
		// Array of documents
		case bsontype.EmbeddedDocument:
			// Pick only the first document
			stillChecking = false
			nested = true
			elemKey := strings.Title(name)
			fmt.Printf("Analysing document %s\n", elemKey)
			nestedFields, err := processDocument(ejvr, opts, elemKey)
			if err != nil {
				fmt.Println(err)
				return nil, snippetResult, fmt.Errorf("error processing nested document for key %q: %w", elemKey, err)
			}
			fmt.Printf("value of nestedFields %s\n", nestedFields)
			retVal = jen.Id(elemKey)
			snippetResult = nestedFields[len(nestedFields)-1]
			fmt.Printf("Value of snippetResult is %s\n", snippetResult)
			fmt.Printf("Value of retval in nested %s\n", retVal)
		default:
			if stillChecking {
				fieldType, _, err := processField(ejvr, opts)
				if err != nil {
					return result, snippetResult, err
				}
				if retVal == nil {
					retVal = fieldType
				} else if retVal.GoString() != fieldType.GoString() {
					stillChecking = false
					retVal = nil
				}
			}
		}
		if nested != true {
			err = ejvr.Skip()
			if err != nil {
				return result, snippetResult, err
			}
		}
	}
	if err != nil && err != bsonrw.ErrEOA {
		return result, snippetResult, err
	}

	if retVal != nil {
		result = jen.Index().Add(retVal)
	}
	return result, snippetResult, nil
}
