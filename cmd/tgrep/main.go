package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"time"

	extractor "github.com/spectrec/moregrep/internal/date/extractor"
	grep "github.com/spectrec/moregrep/internal/date/grep"
	date_profile "github.com/spectrec/moregrep/internal/date/profile"
)

var dateExtractRegexp = flag.String("date-regexp", "", "specify regexp for date extraction")
var dateFormat = flag.String("date-format", "", "specify date format, template: `Jan 2 15:04:05 -0700 MST 2006'")
var dateProfile = flag.String("date-profile", "", "specify date profile name (mescalito|capron|mailloader|mailloader-blob|tarantool|imap|deliveryd|zeptoproxy|msyncd)")

var startDate = flag.String("start-date", "", "specify start date (the same format like in log)")
var endDate = flag.String("end-date", "", "specify end date (the same format like in log)")
var locationName = flag.String("time-zone", "Local", "specify default timezone")

var showLinesWithoutTime = flag.Bool("show-lines-without-date", false, "show lines even when date parse is failed")
var useBinarySearch = flag.Bool("use-binary-search", true, "enable binary search for grep")

var debug = flag.Bool("debug", false, "show debug output")

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %v [<options>] <-|filename>[<-|filename> ...]\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	var profile = date_profile.Search(*dateProfile)
	if *dateProfile != "" && profile == nil {
		fmt.Fprintf(os.Stderr, "unknown date-profile `%v'\n", *dateProfile)
	}

	if profile == nil {
		if *dateExtractRegexp == "" {
			fmt.Fprintf(os.Stderr, "missed -date-regexp/-date-profile\n")
			os.Exit(1)
		}
		dateRegexp, err := regexp.Compile(*dateExtractRegexp)
		if err != nil {
			fmt.Fprintf(os.Stderr, "can't compile regexp `%v': %v\n", *dateExtractRegexp, err)
			os.Exit(1)
		}

		if *dateFormat == "" {
			fmt.Fprintf(os.Stderr, "missed -date-format/-date-profile\n")
			os.Exit(1)
		}
		profile = &date_profile.Profile{
			Regexp: dateRegexp,
			Format: *dateFormat,
		}
	}

	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	location, err := time.LoadLocation(*locationName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "can't parse location `%v': %v\n", *locationName, err)
		os.Exit(1)
	}
	dateExtractor := extractor.NewExtractor(profile.Regexp, profile.Format, location, *debug)

	var startTS, endTS time.Time
	if *startDate != "" {
		v := dateExtractor.Parse(*startDate)
		if v == nil {
			fmt.Fprintf(os.Stderr, "can't parse -start-date (expected format `%v')\n", profile.Format)
			os.Exit(1)
		}
		startTS = *v
	}
	if *endDate != "" {
		v := dateExtractor.Parse(*endDate)
		if v == nil {
			fmt.Fprintf(os.Stderr, "can't parse -end-date (expected format `%v')\n", profile.Format)
			os.Exit(1)
		}
		endTS = *v
	} else {
		endTS = time.Now()
	}

	if startTS.After(endTS) {
		fmt.Fprintf(os.Stderr, "-start-date must be lower or equal to -end-date\n")
		os.Exit(1)
	}

	options := grep.Options{
		StartTime: startTS,
		EndTime:   endTS,

		UseBinSearch:         *useBinarySearch,
		ShowLinesWithoutTime: *showLinesWithoutTime,
	}

	fmt.Fprintf(os.Stderr, "Run grep from `%v' to `%v'\n", startTS, endTS)

	var needStdin bool
	for _, arg := range args {
		if arg == "-" {
			needStdin = true
			continue
		}

		options.Prefix = arg
		if err := doGrep(arg, dateExtractor, options); err != nil {
			fmt.Fprintf(os.Stderr, "grep on `%v' failed: %v\n", arg, err)
		}
	}

	if needStdin {
		options.UseBinSearch = false // can't seek on stdin
		options.Prefix = ""

		doGrep("-", dateExtractor, options)
	}
}

func doGrep(input string, dateExtractor *extractor.DateExtractor, options grep.Options) error {
	var stream io.ReadSeeker

	if input == "-" {
		stream = os.Stdin
	} else {
		file, err := os.Open(input)
		if err != nil {
			return fmt.Errorf("can't open file `%v': %v", input, err)
		}
		defer file.Close()

		stream = file
	}

	return grep.NewGrep(stream, dateExtractor, options).Grep(os.Stdout)
}
