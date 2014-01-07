package bencoding

import (
	"errors"	
	"testing"	
)

// This variable is to handle the case where a package returns a custom error
// such as "strconv.ParseInt: parsing "blabla": invalid syntax".
// In that case, we just expect "some error" but we don't
// really care about the exact error.
var ErrSomeError = errors.New("Some error")

func Test_ParseString(t *testing.T) {
	type StringTest struct {
		input string
		output string
		err error
	}
	
	var stringTests = []StringTest{
		{ "1:a", "a", nil },
		{ "",  "", ErrEmptyInput },
		{ ":abcd", "", ErrInvalidFormat },
		{ "12:123456789 12", "123456789 12", nil },
		{ "12:12345:789 12", "12345:789 12", nil },
		{ "123:abcd", "", ErrInvalidLength },
	}
	
	for _, d := range stringTests {
		output, _, err := parseString([]byte(d.input), 0)
		if err != d.err       { t.Errorf("Expected error '%s', got error '%s'", d.err, err) }
		if output != d.output { t.Errorf("Expected \"%s\", got \"%s\"", d.output, output) }
	}
}

func Test_ParseInt(t *testing.T) {
	type IntTest struct {
		input string
		output int
		err error
	}
	
	var intTests = []IntTest{
		{ "ie", 0, ErrInvalidFormat },
		{ "i", 0, ErrInvalidFormat },
		{ "e", 0, ErrInvalidFormat },
		{ "", 0, ErrEmptyInput },
		{ "iblablae", 0, ErrSomeError },
		{ "i123e", 123, nil },
		{ "i1e", 1, nil },
		{ "i0e", 0, nil },
		{ "i-1e", -1, nil },
		{ "i-123e", -123, nil },
		{ "i-e", 0, ErrSomeError },
	}
	
	for _, d := range intTests {
		output, _, err := parseInt([]byte(d.input), 0)
		if err != nil && d.err == ErrSomeError { err = ErrSomeError }
		if err != d.err       { t.Errorf("Expected error '%s', got error '%s'", d.err, err) }
		if output != d.output { t.Errorf("Expected \"%d\", got \"%d\"", d.output, output) }
	}
}
