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

func compareStringList(anyStringList []*Any, stringList []string) bool {
	if len(anyStringList) != len(stringList) { return false }
	for i, e := range anyStringList {
		s := stringList[i]
		if s != e.AsString { return false }
	}
	return true
}

func compareAny(any1 *Any, any2 *Any) bool {
	if any1.Type != any2.Type { return false }
	if any1.Type == String { return any1.AsString == any2.AsString }
	if any1.Type == Int { return any1.AsInt == any2.AsInt }
	if any1.Type == List {
		for i, e := range any1.AsList {
			equal := compareAny(e, any2.AsList[i])
			if !equal { return false }
		}
		return true
	}
	panic("Unreachable")
}

func Test_ParseList(t *testing.T) {
	type StringListTest struct {
		input string
		output []string
		err error
	}
	
	var stringListTests = []StringListTest{
		{ "", []string{}, ErrEof },
		{ "l", []string{}, ErrInvalidFormat },
		{ "le", []string{}, nil },
		{ "e", []string{}, ErrInvalidFormat },
		{ "l1e", []string{}, ErrInvalidFormat },
		{ "l1:ae", []string{"a"}, nil },
		{ "l1:a2:abe", []string{"a","ab"}, nil },
		{ "l1:a2:ab3:12e", []string{}, ErrInvalidFormat },
	}
	
	for _, d := range stringListTests {
		output, _, err := parseList([]byte(d.input), 0)
		if err != nil && d.err == ErrSomeError { err = ErrSomeError }
		if err != d.err                         { t.Errorf("Expected error '%s', got error '%s' for input '%s'", d.err, err, d.input) }
		if !compareStringList(output, d.output) { t.Errorf("Expected \"%s\", got \"%s\"", d.output, output) }
	}

	type MixListTest struct {
		input string
		output *Any
		err error
	}
	
	var mixListTests = []MixListTest{}
	
	var mixListTest MixListTest
	mixListTest.input = "li123e3:abcl1:x2:yyee"
	mixListTest.output = newAnyList([]*Any{
		newAnyInt(123),
		newAnyString("abc"),
		newAnyList([]*Any{
			newAnyString("x"),
			newAnyString("yy"),
		}),
	})
	mixListTests = append(mixListTests, mixListTest)
	
	for _, d := range mixListTests {
		output, _, err := parseList([]byte(d.input), 0)
		if err != nil && d.err == ErrSomeError  { err = ErrSomeError }
		if err != d.err                         { t.Errorf("Expected error '%s', got error '%s' for input '%s'", d.err, err, d.input) }
		if !compareAny(newAnyList(output), d.output) { t.Errorf("Expected \"%s\", got \"%s\"", d.output, output) }
	}
}
