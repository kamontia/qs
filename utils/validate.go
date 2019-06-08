package utils

import (
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/labstack/gommon/log"
)

func RangeValidation(headMax int, beginNumber int) bool {
	var result bool
	if beginNumber > headMax {
		log.Debug("QS cannot rebase out of range.")
		result = false
	} else if beginNumber == headMax {
		log.Debug("The first commit is included in the specified range. If necessary, please rebase with --root option manually.")
		result = false
	} else {
		log.Debug("Specified range is OK.")
		result = true
	}
	return result
}

func NeedsChangeMessage(target int, begin int, end int) bool {
	if target <= begin && target >= end {
		return true
	} else {
		return false
	}
}

func Validate(n string) bool {
	r := regexp.MustCompile(`^[0-9]+$|^[0-9]+..[0-9]+$`)
	return r.MatchString(n)
}

/*
	$ qs -n 1..4
	HEAD~0:
	HEAD~1:  <- end
	HEAD~2:
	HEAD~3:
	HEAD~4:  <- start
	HEAD~5:
*/
func PickupSquashRange(num string) (int, int) {
	/* TODO: Check error strictly */
	var error error
	var bn int
	var en int

	if strings.Contains(num, "..") {
		/* Specify the range you aggregate */
		rangeArray := strings.Split(num, "..")
		bn, error = strconv.Atoi(rangeArray[0])
		if error != nil {
			log.Error(error)
			os.Exit(1)
		}
		en, error = strconv.Atoi(rangeArray[1])
		if error != nil {
			log.Error(error)
			os.Exit(1)
		}
		if bn < en {
			tmp := bn
			bn = en
			en = tmp
		}
	} else {
		/* Specify the one parameter you aggregate */
		en = 0
		bn, error = strconv.Atoi(num)
		if error != nil {
			log.Error(error)
		}
	}
	return bn, en
}
