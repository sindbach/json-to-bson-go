package simplejson

import (
	"fmt"
	"strings"

	. "github.com/dave/jennifer/jen"
	"github.com/iancoleman/orderedmap"
	"github.com/sindbach/json-to-bson-go/jsonutil"
	"github.com/sindbach/json-to-bson-go/options"
)

func Convert(jsonStr []byte, opts *options.Options) (string, error) {
	if opts == nil {
		opts = options.NewOptions()
	}

	input, err := jsonutil.Unmarshal(jsonStr)
	if err != nil {
		return "", err
	}

	fields, err := convertMapToFields(input, opts)
	if err != nil {
		return "", err
	}

	output := NewFile("main")
	output.Type().Id(opts.StructName()).Struct(fields...)
	return output.GoString(), nil
}

func convertMapToFields(input *orderedmap.OrderedMap, opts *options.Options) ([]Code, error) {
	var fields []Code
	for _, key := range input.Keys() {
		val, _ := input.Get(key)
		structTags := []string{key}

		elem := Id(strings.Title(key))
		switch converted := val.(type) {
		case string:
			elem.Add(String())
		case bool:
			elem.Add(Bool())
		case float64:
			if !opts.MinimizeIntegerSize() {
				elem.Add(Float64())
				break
			}

			canTruncate := true
			switch {
			case float64(int32(converted)) == converted:
				elem.Add(Int32())
			case float64(int64(converted)) == converted:
				elem.Add(Int64())
			default:
				canTruncate = false
				elem.Add(Float64())
			}

			if canTruncate && opts.TruncateIntegers() {
				structTags = append(structTags, "truncate")
			}
		case []interface{}:
			elem.Add(Index().Interface())
			structTags = append(structTags, "omitempty")
		case orderedmap.OrderedMap:
			nestedFields, err := convertMapToFields(&converted, opts)
			if err != nil {
				return nil, fmt.Errorf("error processing nested document for key %q: %w", key, err)
			}

			elem.Add(Struct(nestedFields...))
		default:
			return nil, fmt.Errorf("value for key %q has unrecognized type %T", key, val)
		}

		tagsString := strings.Join(structTags, ",")
		elem.Tag(map[string]string{"bson": tagsString})
		fields = append(fields, elem)
	}

	return fields, nil
}
