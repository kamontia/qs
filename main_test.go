package main

import (
	"testing"

	"github.com/kamontia/qs/utils"
)

func Test_pickupSquashRange(t *testing.T) {
	begin, end := utils.PickupSquashRange("2")
	if begin != 2 || end != 0 {
		t.Errorf("begin: %v\nend: %v", begin, end)
	}

	begin, end = utils.PickupSquashRange("2..4")
	if begin != 4 || end != 2 {
		t.Errorf("begin: %v\nend: %v", begin, end)
	}

	begin, end = utils.PickupSquashRange("4..2")
	if begin != 4 || end != 2 {
		t.Errorf("begin: %v\nend: %v", begin, end)
	}
}
