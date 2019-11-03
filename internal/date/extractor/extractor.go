package date_extractor

import (
	"fmt"
	"regexp"
	"time"
)

type DateExtractor struct {
	reg        *regexp.Regexp
	dateFormat string

	location *time.Location

	debug bool
}

func NewExtractor(reg *regexp.Regexp, dateFormat string, location *time.Location, debug bool) *DateExtractor {
	return &DateExtractor{reg: reg, dateFormat: dateFormat, location: location, debug: debug}
}

func (e *DateExtractor) Parse(date string) *time.Time {
	t, err := time.ParseInLocation(e.dateFormat, date, e.location)
	if err != nil {
		if e.debug {
			fmt.Printf("failed to parse date `%v': `%v'\n", date, err)
		}

		return nil
	}

	if e.debug {
		fmt.Printf("extract date: %v (%v)\n", t, date)
	}

	return &t
}

func (e *DateExtractor) Extract(line string) *time.Time {
	ret := e.reg.FindStringSubmatch(line)
	if len(ret) != 2 {
		if e.debug {
			fmt.Printf("can't extract date from `%v', got: %#v\n", line, ret)
		}

		return nil
	}

	return e.Parse(ret[1])
}
