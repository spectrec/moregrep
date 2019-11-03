package time_range_grep

import (
	"bufio"
	"fmt"
	"io"
	"time"

	extractor "github.com/spectrec/moregrep/internal/date/extractor"
)

type Options struct {
	Prefix string

	StartTime time.Time
	EndTime   time.Time

	UseBinSearch         bool
	ShowLinesWithoutTime bool
}

type TimeRangeGrep struct {
	stream  io.ReadSeeker
	options Options

	dateExtractor *extractor.DateExtractor
}

func NewGrep(stream io.ReadSeeker, dateExtractor *extractor.DateExtractor, options Options) *TimeRangeGrep {
	return &TimeRangeGrep{
		stream:  stream,
		options: options,

		dateExtractor: dateExtractor,
	}
}

func (g *TimeRangeGrep) searchStart() error {
	var startOff int64 = 0

	endOff, err := g.stream.Seek(0, io.SeekEnd)
	if err != nil {
		return fmt.Errorf("can't detect end offset: %v", err)
	}

search_off:
	for {
		var curOff = (startOff + endOff) / 2

		_, err := g.stream.Seek(curOff, io.SeekStart)
		if err != nil {
			return fmt.Errorf("can't seek to `%v' offset: %v", curOff, err)
		}

		reader := bufio.NewReader(g.stream)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					break search_off
				}

				return fmt.Errorf("can't read next line: %v", err)
			}

			if startOff+int64(len(line)) >= endOff {
				break search_off
			}

			t := g.dateExtractor.Extract(line)
			if t == nil {
				// Assume that we have jumped to the middle of the
				//  line, so just try next line
				continue
			}

			if t.Before(g.options.StartTime) {
				startOff = curOff
				break
			}

			if t.After(g.options.StartTime) || t.Equal(g.options.StartTime) {
				endOff = curOff
				break
			}
		}
	}

	// Seek to the last checked offset again (we have read one line starting from it)
	if _, err = g.stream.Seek(startOff, io.SeekStart); err != nil {
		return fmt.Errorf("can't seek to the found offset: %v", err)
	}

	return nil
}

func (g *TimeRangeGrep) Grep(out io.Writer) error {
	var prefix = g.options.Prefix
	var colon = ""

	if prefix != "" {
		colon = ":"
	}

	if g.options.UseBinSearch {
		if err := g.searchStart(); err != nil {
			return fmt.Errorf("can't find start: %v", err)
		}
	}

	reader := bufio.NewReader(g.stream)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}

			return fmt.Errorf("can't read next line: %v", err)
		}

		t := g.dateExtractor.Extract(line)
		if t == nil {
			if !g.options.ShowLinesWithoutTime {
				continue
			}
		} else if t.Before(g.options.StartTime) {
			// Assume that this is line just before required one
			continue
		} else if t.After(g.options.EndTime) {
			if !g.options.UseBinSearch {
				// It is ok only for fullscan mode
				continue
			}

			return nil
		}

		if _, err := fmt.Fprintf(out, "%s%s%s", prefix, colon, []byte(line)); err != nil {
			return fmt.Errorf("failed to write to result stream: %v", err)
		}
	}

	return nil
}
