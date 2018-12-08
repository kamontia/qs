package main

import "testing"

func Test_needsChangeMessage_returnTrue(t *testing.T) {
	needed := needsChangeMessage(3, 4, 2)
	actual := needed
	expected := true
	if actual != expected {
		t.Errorf("got: %v\nwant: %v", actual, expected)
	}
}

func Test_needsChangeMessage_returnFalse(t *testing.T) {
	needed := needsChangeMessage(5, 4, 2)
	actual := needed
	expected := false
	if actual != expected {
		t.Errorf("got: %v\nwant: %v", actual, expected)
	}
}

func Test_pickupSquashRange(t *testing.T) {
	begin, end := pickupSquashRange("2")
	if begin != 2 || end != 0 {
		t.Errorf("begin: %v\nend: %v", begin, end)
	}

	begin, end = pickupSquashRange("2..4")
	if begin != 4 || end != 2 {
		t.Errorf("begin: %v\nend: %v", begin, end)
	}

	begin, end = pickupSquashRange("4..2")
	if begin != 4 || end != 2 {
		t.Errorf("begin: %v\nend: %v", begin, end)
	}
}

func Test_rangeVaridation(t *testing.T) {
	result := rangeValidation(3, 1, 2)
	if !result {
		t.Error("range validation is failed")
	}
}
