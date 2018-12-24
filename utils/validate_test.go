package utils

import "testing"

func Test_RangeVaridation(t *testing.T) {
	result := RangeValidation(3, 1)
	if !result {
		t.Errorf("range validation is failed. return %v", result)
	}

	result = RangeValidation(1, 3)
	if result {
		t.Errorf("range validation is failed. return %v", result)
	}

	result = RangeValidation(1, 1)
	if result {
		t.Errorf("range validation is failed. return %v", result)
	}
}
func Test_NeedsChangeMessage_returnTrue(t *testing.T) {
	needed := NeedsChangeMessage(3, 4, 2)
	actual := needed
	expected := true
	if actual != expected {
		t.Errorf("got: %v\nwant: %v", actual, expected)
	}
}

func Test_NeedsChangeMessage_returnFalse(t *testing.T) {
	needed := NeedsChangeMessage(5, 4, 2)
	actual := needed
	expected := false
	if actual != expected {
		t.Errorf("got: %v\nwant: %v", actual, expected)
	}
}

func Test_Validate(t *testing.T) {
	isNum := Validate("2")
	if !isNum {
		t.Errorf("Validate return %v", isNum)
	}
	isNum = Validate("2..4")
	if !isNum {
		t.Errorf("Validate return %v", isNum)
	}
	isNum = Validate("hoge")
	if isNum {
		t.Errorf("Validate return %v", isNum)
	}
}
