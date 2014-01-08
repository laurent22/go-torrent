package bencoding

import (
	"errors"
	"sort"
	"strconv"
)

var ErrEmptyInput = errors.New("input cannot be empty")
var ErrInvalidFormat = errors.New("invalid format")
var ErrInvalidLength = errors.New("invalid length")
var ErrUnsupportedType = errors.New("unsupported type")
var ErrEof = errors.New("end of stream")

const (
	String = 1
	Int = 2
	List = 3
	Dictionary = 4
)

type Any struct {
	Type int
	AsString string
	AsInt int
	AsList []*Any
	AsDictionary map[string]*Any
}

func dumpIndentSpaces(count int) string {
	output := ""
	for i := 0; i < count; i++ { output += "    " }
	return output
}

func (this *Any) dump(indent int) string {
	if this.Type == String {
		return "\"" + this.AsString + "\""
	}

	if this.Type == Int {
		return strconv.Itoa(this.AsInt)
	}
	
	if this.Type == List {
		output := "[\n"
		for i, e := range this.AsList {
			output += dumpIndentSpaces(indent + 1) + strconv.Itoa(i) + ": "
			output += e.dump(indent + 1) + "\n"
		}
		output += dumpIndentSpaces(indent) + "]"
		return output
	}
	
	if this.Type == Dictionary {
		output := "{\n"
		for k, e := range this.AsDictionary {
			output += dumpIndentSpaces(indent + 1) + k + ": "
			output += e.dump(indent + 1) + "\n"
		}
		output += dumpIndentSpaces(indent) + "}"
		return output
	}
	
	panic("unreachable")
}

func (this *Any) Dump() string {
	return this.dump(0)
}

const (
	stStarting = 0
)

func newAnyString(s string) *Any {
	output := new(Any)
	output.Type = String
	output.AsString = s
	return output
}

func newAnyInt(s int) *Any {
	output := new(Any)
	output.Type = Int
	output.AsInt = s
	return output
}

func newAnyList(s []*Any) *Any {
	output := new(Any)
	output.Type = List
	output.AsList = s
	return output
}

func newAnyDictionary(s map[string]*Any) *Any {
	output := new(Any)
	output.Type = Dictionary
	output.AsDictionary = s
	return output
}

func byteIndex(input []byte, n byte, startIndex int) int {
	for i := startIndex; i < len(input); i++ {
		b := input[i]
		if b == n {
			return i
		}
	}
	return -1
}

func decodeString(input []byte, index int) (string, int, error) {
	if index >= len(input) { return "", index, ErrEof }
	colonIndex := byteIndex(input, ':', index)
	if colonIndex <= 0 { return "", index, ErrInvalidFormat }
	stringLength, err := strconv.Atoi(string(input[index:colonIndex]))
	if err != nil { return "", colonIndex + 1, err }
	if colonIndex + stringLength >= len(input) { return "", colonIndex + 1, ErrInvalidLength }
	output := input[colonIndex + 1 : colonIndex + 1 + stringLength]
	return string(output), colonIndex + stringLength + 1, nil
}

func decodeInt(input []byte, index int) (int, int, error) {
	if index >= len(input) { return 0, index, ErrEof }
	if input[index] != 'i' { return 0, index, ErrInvalidFormat }
	endIndex := byteIndex(input, 'e', index + 1)
	if endIndex <= index + 1 { return 0, index + 1, ErrInvalidFormat }
	output, err := strconv.Atoi(string(input[index + 1 : endIndex]))
	if err != nil { return 0, index + 1, err }
	return output, endIndex + 1, nil
}

func decodeList(input []byte, index int) ([]*Any, int, error) {
	if index >= len(input) { return []*Any{}, index, ErrEof }
	if input[index] != 'l' { return []*Any{}, index, ErrInvalidFormat }

	var output []*Any
	var i int
	for i = index + 1; i < len(input); i++ {
		if input[i] == 'e' {
			index = i + 1
			return output, index, nil
		}
		item, newIndex, err := decodeNext(input, i)
		if err != nil { return output, i, err }
		i = newIndex - 1 // Decrement since it's going to be incremented in the for statement
		output = append(output, item)
	}
	return []*Any{}, i, ErrInvalidFormat // Didn't find 'e' tag
}

func decodeDictionary(input []byte, index int) (map[string]*Any, int, error) {
	if index >= len(input) { return map[string]*Any{}, index, ErrEof }
	if input[index] != 'd' { return map[string]*Any{}, index, ErrInvalidFormat }

	output := make(map[string]*Any)
	var i int
	for i = index + 1; i < len(input); i++ {
		if input[i] == 'e' {
			index = i + 1
			return output, index, nil
		}

		key, newIndex, err := decodeString(input, i)
		if err != nil { return map[string]*Any{}, newIndex, err }
		i = newIndex

		value, newIndex, err := decodeNext(input, i)		
		if err != nil { return map[string]*Any{}, newIndex, err }
		i = newIndex - 1 // Decrement since it's going to be incremented in the for statement
		
		output[key] = value
	}
	return map[string]*Any{}, i, ErrInvalidFormat // Didn't find 'e' tag
}

func decodeNext(input []byte, index int) (*Any, int, error) {	
	if index >= len(input) { return nil, index, ErrEof }
	b := input[index]
	switch {

		case b >= '0' && b <= '9':
			
			s, index, err := decodeString(input, index)
			if err != nil { return nil, index, err }
			return newAnyString(s), index, nil

		case b == 'i':

			i, index, err := decodeInt(input, index)
			if err != nil { return nil, index, err }
			return newAnyInt(i), index, nil
			
		case b == 'l':
			
			l, index, err := decodeList(input, index)
			if err != nil { return nil, index, err }
			return newAnyList(l), index, nil

		case b == 'd':

			d, index, err := decodeDictionary(input, index)
			if err != nil { return nil, index, err }
			return newAnyDictionary(d), index, nil
	}

	return nil, index, ErrUnsupportedType
}

func appendBytes(dest []byte, source []byte) []byte {
	output := dest
	for _, b := range source {
		output = append(output, b)
	}
	return output
}

func Decode(input []byte) (*Any, error) {
	output, _, err := decodeNext(input, 0)
	return output, err
}

func Encode(any *Any) ([]byte, error) {
	if any.Type == String {
		return []byte(strconv.Itoa(len(any.AsString)) + ":" + any.AsString), nil
	}
	
	if any.Type == Int {
		return []byte("i" + strconv.Itoa(any.AsInt) + "e"), nil
	}
	
	if any.Type == List {
		output := []byte{'l'}
		for _, e := range any.AsList {
			encoded, err := Encode(e)
			if err != nil { return []byte{}, nil }
			output = appendBytes(output, encoded)
		}
		output = append(output, 'e')
		return output, nil
	}
	
	if any.Type == Dictionary {
		// The keys must be sorted first
		var keys []string
		for key, _ := range any.AsDictionary {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		output := []byte{'d'}
		for _, key := range keys {
			e := any.AsDictionary[key]
			encodedKey, err := Encode(newAnyString(key))
			if err != nil { return []byte{}, nil }
			output = appendBytes(output, encodedKey)
			encoded, err := Encode(e)
			if err != nil { return []byte{}, nil }
			output = appendBytes(output, encoded)
		}
		output = append(output, 'e')
		return output, nil
	}
	
	panic("unreachable")
}