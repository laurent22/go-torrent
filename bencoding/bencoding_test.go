package bencoding

import (
	"errors"	
	"testing"	
)

// This variable is to handle the case where a package returns a custom error
// such as "strconv.DecodeInt: parsing "blabla": invalid syntax".
// In that case, we just expect "some error" but we don't
// really care about the exact error.
var ErrSomeError = errors.New("Some error")

func Test_DecodeString(t *testing.T) {
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
		output, _, err := decodeString([]byte(d.input), 0)
		if err != d.err       { t.Errorf("Expected error '%s', got error '%s'", d.err, err) }
		if output != d.output { t.Errorf("Expected \"%s\", got \"%s\"", d.output, output) }
	}
}

func Test_DecodeInt(t *testing.T) {
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
		output, _, err := decodeInt([]byte(d.input), 0)
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
	if any1 == nil && any2 != nil { return false }
	if any1 != nil && any2 == nil { return false }
	if any1 == nil && any2 == nil { return true }
	if any1.Type != any2.Type { return false }

	if any1.Type == String { return any1.AsString == any2.AsString }

	if any1.Type == Int { return any1.AsInt == any2.AsInt }

	if any1.Type == List {
		if len(any1.AsList) != len(any2.AsList) { return false }
		for i, e := range any1.AsList {
			equal := compareAny(e, any2.AsList[i])
			if !equal { return false }
		}
		return true
	}

	if any1.Type == Dictionary {
		if len(any1.AsDictionary) != len(any2.AsDictionary) { return false }
		for k, e := range any1.AsDictionary {
			equal := compareAny(e, any2.AsDictionary[k])
			if !equal { return false }
		}
		return true
	}

	panic("Unreachable")
}

func Test_DecodeList(t *testing.T) {
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
		output, _, err := decodeList([]byte(d.input), 0)
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
		output, _, err := decodeList([]byte(d.input), 0)
		if err != nil && d.err == ErrSomeError  { err = ErrSomeError }
		if err != d.err { t.Errorf("Expected error '%s', got error '%s' for input '%s'", d.err, err, d.input) }
		if !compareAny(newAnyList(output), d.output) { t.Errorf("Expected \"%s\", got \"%s\"", d.output, output) }
	}
}

func Test_DecodeDictionary(t *testing.T) {
	type MixListTest struct {
		input string
		output map[string]*Any
		err error
	}
	
	var mixListTests = []MixListTest{}
	
	var d MixListTest
	
	d.input = "di123e3:abcl1:x2:yyee"
	d.output = map[string]*Any{}
	d.err = ErrSomeError
	mixListTests = append(mixListTests, d)
	
	d.input = ""
	d.output = map[string]*Any{}
	d.err = ErrEof
	mixListTests = append(mixListTests, d)
	
	d.input = "d"
	d.output = map[string]*Any{}
	d.err = ErrInvalidFormat
	mixListTests = append(mixListTests, d)
	
	d.input = "d3:key4AAAAe"
	d.output = map[string]*Any{}
	d.err = ErrInvalidFormat
	mixListTests = append(mixListTests, d)

	d.input = "d3:key4:AAAAe"
	d.output = map[string]*Any{
		"key": newAnyString("AAAA"),
	}
	d.err = nil
	mixListTests = append(mixListTests, d)

	d.input = "d3:key4:AAAA4:key2d2:XXi123e3:XXXli123ei456eeee"
	d.output = map[string]*Any{
		"key": newAnyString("AAAA"),
		"key2": newAnyDictionary(map[string]*Any{
			"XX": newAnyInt(123),
			"XXX": newAnyList([]*Any{
				newAnyInt(123),
				newAnyInt(456),
			}),
		}),
	}
	d.err = nil
	mixListTests = append(mixListTests, d)
	
	for _, d := range mixListTests {
		output, index, err := decodeDictionary([]byte(d.input), 0)
		if err != nil && d.err == ErrSomeError  { err = ErrSomeError }
		if err != d.err { t.Errorf("Expected error '%s', got error '%s' for input '%s' at index %d", d.err, err, d.input, index) }
		if !compareAny(newAnyDictionary(output), newAnyDictionary(d.output)) { t.Errorf("Expected \"%s\", got \"%s\" at index %d", d.output, output, index) }
	}
}

func Test_Decode(t *testing.T) {
	{
		output, err := Decode([]byte("4:abcd"))
		if output.Type != String { t.Errorf("Expected string type, got %d", output.Type) }
		if output.AsString != "abcd" { t.Errorf("Expected 'abcd', got %s", output.AsString) }
		if err != nil { t.Errorf("Got error", err) }
	}
	
	{
		output, err := Decode([]byte("i1234e"))
		if output.Type != Int { t.Errorf("Expected int type, got %d", output.Type) }
		if output.AsInt != 1234 { t.Errorf("Expected 1234, got %d", output.AsInt) }
		if err != nil { t.Errorf("Got error", err) }
	}
}

func Test_Encode(t *testing.T) {
	var stringTests = []string{
		"d3:key4:AAAA4:key2d2:XXi123e3:XXXli123ei456eeee",
		"d3:key4:AAAAe",
		"li123e3:abcl1:x2:yyee",
		"l1:a2:abe",
		"12:12345:789 12",
		"3:abc",
		"i123e",
		"i1e",
		"i0e",
		"i-1e",
		"i-123e",
	}
	
	for _, s := range stringTests {
		decoded, err := Decode([]byte(s))
		if err != nil {
			t.Fatal("Invalid input string:", s)
		}
		encoded, err := Encode(decoded)
		if err != nil {
			t.Error("Expected no error, got", err)
		}
		if string(encoded) != s {
			t.Errorf("Expected '%s', got '%s'", s, string(encoded))
		}
	}
}