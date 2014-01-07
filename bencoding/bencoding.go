package bencoding

import (
	"errors"
	"strconv"
)

var ErrEmptyInput = errors.New("input cannot be empty")
var ErrInvalidFormat = errors.New("invalid format")
var ErrInvalidLength = errors.New("invalid length")
var ErrUnsupportedType = errors.New("unsupported type")

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
	if len(input) == 0 { return "", index, ErrEmptyInput }
	colonIndex := byteIndex(input, ':', index)
	if colonIndex <= 0 { return "", index, ErrInvalidFormat }
	stringLength, err := strconv.Atoi(string(input[index:colonIndex]))
	if err != nil { return "", index, err }
	output := input[colonIndex+1:]
	if len(output) != stringLength { return "", index, ErrInvalidLength }
	return string(output), index + colonIndex + stringLength + 1, nil
}

func parseInt(input []byte, index int) (int, int, error) {
	if len(input) == 0 { return 0, index, ErrEmptyInput }
	if input[index] != 'i' { return 0, index, ErrInvalidFormat }
	endIndex := byteIndex(input, 'e', index + 1)
	if endIndex <= index + 1 { return 0, index, ErrInvalidFormat }
	output, err := strconv.Atoi(string(input[index + 1 : endIndex]))
	if err != nil { return 0, index, err }
	return output, endIndex + 1, nil
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