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
		{ "", "", ErrEof },
		{ "0:", "", nil },
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
		{ "", 0, ErrEof },
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

func compareStringList(anyStringList []Any, stringList []string) bool {
	if len(anyStringList) != len(stringList) { return false }
	for i, e := range anyStringList {
		s := stringList[i]
		if s != e.AsString { return false }
	}
	return true
}

func Test_ParseList(t *testing.T) {
	type StringListTest struct {
		input string
		output []string
		err error
	}
	
	var stringListTest = []StringListTest{
		{ "", []string{}, ErrEof },
		{ "l", []string{}, ErrInvalidFormat },
		{ "le", []string{}, nil },
		{ "e", []string{}, ErrInvalidFormat },
		{ "l1e", []string{}, ErrInvalidFormat },
		{ "l1:ae", []string{"a"}, nil },
		{ "l1:a2:abe", []string{"a","ab"}, nil },
		{ "l1:a2:ab3:12e", []string{}, ErrInvalidFormat },
	}
	
	for _, d := range stringListTest {
		output, _, err := parseList([]byte(d.input), 0)
		if err != nil && d.err == ErrSomeError { err = ErrSomeError }
		if err != d.err                         { t.Errorf("Expected error '%s', got error '%s' for input '%s'", d.err, err, d.input) }
		if !compareStringList(output, d.output) { t.Errorf("Expected \"%s\", got \"%s\"", d.output, output) }
	}
}
