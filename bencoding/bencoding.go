package bencoding

import (
	"errors"
	"fmt"
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
	AsList []Any
	AsDictionary map[string]Any
}

const (
	stStarting = 0
)

func byteIndex(input []byte, n byte, startIndex int) int {
	for i := startIndex; i < len(input); i++ {
		b := input[i]
		if b == n {
			return i
		}
	}
	return -1
}

func parseString(input []byte, index int) (string, int, error) {
	if index >= len(input) { return "", index, ErrEof }
	colonIndex := byteIndex(input, ':', index)
	if colonIndex <= 0 { return "", index, ErrInvalidFormat }
	stringLength, err := strconv.Atoi(string(input[index:colonIndex]))
	if err != nil { return "", index, err }
	if colonIndex + stringLength >= len(input) { return "", index, ErrInvalidLength }
	output := input[colonIndex + 1 : colonIndex + 1 + stringLength]
	return string(output), colonIndex + stringLength + 1, nil
}

// TODO: handle len(input) == 0
func parseInt(input []byte, index int) (int, int, error) {
	if index >= len(input) { return 0, index, ErrEof }
	if input[index] != 'i' { return 0, index, ErrInvalidFormat }
	endIndex := byteIndex(input, 'e', index + 1)
	if endIndex <= index + 1 { return 0, index, ErrInvalidFormat }
	output, err := strconv.Atoi(string(input[index + 1 : endIndex]))
	if err != nil { return 0, index, err }
	return output, endIndex + 1, nil
}

func parseList(input []byte, index int) ([]Any, int, error) {
	_ = fmt.Println

	if index >= len(input) { return []Any{}, index, ErrEof }
	if input[index] != 'l' { return []Any{}, index, ErrInvalidFormat }

	var output []Any
	for i := index + 1; i < len(input); i++ {
		if input[i] == 'e' {
			index = i + 1
			return output, index, nil
		}
		item, newIndex, err := parseNext(input, i)
		i = newIndex - 1 // Decrement since it's going to be incremented in the for statement
		if err != nil { return output, i, err }
		output = append(output, item)
	}
	return []Any{}, index, ErrInvalidFormat // Didn't find 'e' tag
}

func parseDictionary(input []byte, index int) (map[string]Any, int, error) {
	var output map[string]Any
	return output, index, nil
}

func parseNext(input []byte, index int) (Any, int, error) {	
	var output Any
	if index >= len(input) { return output, index, ErrEof }
	b := input[index]
	switch {

		case b >= '0' && b <= '9':
			
			s, index, err := parseString(input, index)
			if err != nil { return output, index, err }
			output.AsString = s
			output.Type = String
			return output, index, nil

		case b == 'i':

			i, index, err := parseInt(input, index)
			if err != nil { return output, index, err }
			output.AsInt = i
			output.Type = Int
			return output, index, nil
			
		case b == 'l':
			
			l, index, err := parseList(input, index)
			if err != nil { return output, index, err }
			output.AsList = l
			output.Type = List
			return output, index, nil

		case b == 'd':

			d, index, err := parseDictionary(input, index)
			if err != nil { return output, index, err }
			output.AsDictionary = d
			output.Type = List
			return output, index, nil
	}

	return output, index, ErrUnsupportedType
}

func Parse(input []byte) (Any, error) {
	var output Any
	if len(input) == 0 { return output, ErrEmptyInput }
	
	b := input[0]
	switch {

		case b >= '0' && b <= '9':

			s, _, err := parseString(input, 0)
			if err != nil { return output, err }
			output.AsString = s
			output.Type = String

		case b == 'i':

			i, _, err := parseInt(input, 0)
			if err != nil { return output, err }
			output.AsInt = i
			output.Type = Int

		case b == 'l':

			// list

		case b == 'd':

			//dictionary
			
		default:
			
			return output, ErrUnsupportedType

	}
	
	return output, nil
}