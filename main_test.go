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
