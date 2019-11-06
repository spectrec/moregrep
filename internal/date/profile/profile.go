package date_profile

import (
	"regexp"
)

type Profile struct {
	Regexp *regexp.Regexp
	Format string
}

var profiles = map[string]Profile{
	// mescalito-worker._disk874.1[25972 03.11.2019 16:51:11.882525]
	"mescalito": Profile{
		Regexp: regexp.MustCompile(`(?:\[\d+ (\d+\.\d+\.\d+ \d+:\d+:\d+))`),
		Format: `02.01.2006 15:04:05`,
	},

	// 2019-11-03T16:51:51.000 mclost1523-1 zeptoproxy[23939]: 2019/11/03 16:51:51.767770
	"zeptoproxy": Profile{
		Regexp: regexp.MustCompile(`(?:zeptoproxy.+?: (\d+/\d+/\d+ \d+:\d+:\d+))`),
		Format: `2006/01/02 15:04:05`,
	},

	// msyncd[8312 03.11.2019 19:48:35.000000]
	"msyncd": Profile{
		Regexp: regexp.MustCompile(`(?:msyncd.+?: (\d+\.\d+\.\d+ \d+:\d+:\d+))`),
		Format: `02.01.2006 15:04:05`,
	},

	// [4455 03.11.2019 01:02:40.070113]
	"capron": Profile{
		Regexp: regexp.MustCompile(`(?:\[\d+ (\d+\.\d+\.\d+ \d+:\d+:\d+))`),
		Format: `02.01.2006 15:04:05`,
	},

	// 2019-11-03 18:00:10.440
	"deliveryd": Profile{
		Regexp: regexp.MustCompile(`(\d+-\d+-\d+ \d+:\d+:\d+)`),
		Format: `2006-01-02 15:04:05`,
	},

	// 2019-11-03 18:00:10.440
	"mailloader": Profile{
		Regexp: regexp.MustCompile(`(\d+-\d+-\d+ \d+:\d+:\d+)`),
		Format: `2006-01-02 15:04:05`,
	},

	// Nov  3 17:07:01.847
	"mailloader-blob": Profile{
		Regexp: regexp.MustCompile(`(\S+\s+\d+ \d+:\d+:\d+)`),
		Format: `Jan _2 15:04:05`,
	},

	// <hostname> 1572788668:003 Nov  3 16:44:28
	"imap": Profile{
		Regexp: regexp.MustCompile(`(\S+\s+\d+ \d+:\d+:\d+)`),
		Format: `Jan _2 15:04:05`,
	},

	// tarantool_xtaz_101: 2019-11-03 00:00:04.981
	"tarantool": Profile{
		Regexp: regexp.MustCompile(`(?:tarantool_xtaz.+?: (\d+-\d+-\d+ \d+:\d+:\d+))`),
		Format: `2006-01-02 15:04:05`,
	},
}

func Search(name string) *Profile {
	profile, exist := profiles[name]
	if !exist {
		return nil
	}

	return &profile
}
