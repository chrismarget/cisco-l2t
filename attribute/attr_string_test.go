package attribute

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
	"unicode"
)

func TestStringAttribute_String(t *testing.T) {
	// build the character set to be used when generating strings
	var stringRunes []rune
	for i := 0; i <= unicode.MaxASCII; i++ {
		if unicode.IsPrint(rune(i)) {
			stringRunes = append(stringRunes, rune(i))
		}
	}

	// stringStringTestData map contains test attribute data
	// ([]byte{"f","o","o",stringTerminator}) and the expected
	// Attribute.String() result ("foo")
	stringStringTestData := make(map[string]string)

	// first string to test is empty string
	stringStringTestData[string(stringTerminator)] = string("")

	// next string to test is maximum size random string
	runeSlice := make([]rune, maxStringWithoutTerminator)
	for c := range runeSlice {
		runeSlice[c] = stringRunes[rand.Intn(len(stringRunes))]
	}
	stringStringTestData[string(runeSlice)+string(stringTerminator)] = string(runeSlice)

	// Let's add 98 more random strings of random length
	rand.Seed(time.Now().UnixNano())
	for i := 1; i <= 98; i++ {
		strlen := rand.Intn(maxStringWithoutTerminator)
		runeSlice := make([]rune, strlen)
		// make a slice of random "stringy" runes, "i" bytes long
		for c := range runeSlice {
			runeSlice[c] = stringRunes[rand.Intn(len(stringRunes))]
		}
		stringStringTestData[string(runeSlice)+string(stringTerminator)] = string(runeSlice)
	}

	for _, stringAttrType := range getAttrsByCategory(stringCategory) {
		for data, expected := range stringStringTestData {
			testAttr := stringAttribute{
				attrType: stringAttrType,
				attrData: []byte(data),
			}
			result := testAttr.String()
			if result != expected {
				t.Fatalf("expected %s, got %s", expected, result)
			}
		}
	}
}

func TestStringAttribute_Validate_WithGoodData(t *testing.T) {
	// build the character set to be used when generating strings
	var stringRunes []rune
	for i := 0; i <= unicode.MaxASCII; i++ {
		if unicode.IsPrint(rune(i)) {
			stringRunes = append(stringRunes, rune(i))
		}
	}

	// first example of good data is an empy string (terminator only)
	goodData := [][]byte{[]byte{stringTerminator}}

	// Now lets build 3 goodData entries for each allowed character.
	// Lengths will be 1, <random>, and maxStringWithoutTerminator.
	for _, c := range stringRunes {
		testData := make([]byte, maxStringWithoutTerminator)
		for i := 0; i < maxStringWithoutTerminator; i++ {
			testData[i] = byte(c)
		}
		short := 1
		medium := rand.Intn(maxStringWithoutTerminator)
		//medium := 5
		long := maxStringWithoutTerminator

		var addMe string

		addMe = string(testData[0:short]) + string(stringTerminator)
		goodData = append(goodData, []byte(addMe))
		addMe = string(testData[0:medium]) + string(stringTerminator)
		goodData = append(goodData, []byte(addMe))
		addMe = string(testData[0:long]) + string(stringTerminator)
		goodData = append(goodData, []byte(addMe))
	}

	for _, stringAttrType := range getAttrsByCategory(stringCategory) {
		for _, testData := range goodData {
			testAttr := stringAttribute{
				attrType: stringAttrType,
				attrData: testData,
			}
			err := testAttr.Validate()
			if err != nil {
				t.Fatalf(err.Error()+"\n"+"Supposed good data %s produced error for %s.",
					fmt.Sprintf("%v", []byte(testAttr.attrData)), attrTypeString[stringAttrType])
			}
		}
	}
}

func TestStringAttribute_Validate_WithBadData(t *testing.T) {
	// build the character set to be used when generating bogus strings
	var badStringRunes []rune
	for i := 0; i <= unicode.MaxASCII; i++ {
		if !unicode.IsPrint(rune(i)) {
			badStringRunes = append(badStringRunes, rune(i))
		}
	}

	badData := [][]byte{
		nil,                  // unterminated
		[]byte{},             // unterminated
		[]byte{65},           // unterminated
		[]byte{65, 0, 0},     //embedded terminator
		[]byte{65, 0, 65},    //embedded terminator
		[]byte{65, 0, 65, 0}, //embedded terminator
	}

	// Add some properly terminated strings containing bogus characters
	for _, bsr := range badStringRunes {
		badString := []byte(string(bsr) + string(stringTerminator))
		badData = append(badData, badString)
	}

	for _, stringAttrType := range getAttrsByCategory(stringCategory) {
		for _, testData := range badData {
			testAttr := stringAttribute{
				attrType: stringAttrType,
				attrData: testData,
			}

			err := testAttr.Validate()
			if err == nil {
				t.Fatalf("Bad data %s in %s did not error.",
					fmt.Sprintf("%v", []byte(testAttr.attrData)), attrTypeString[stringAttrType])
			}
		}
	}
}
