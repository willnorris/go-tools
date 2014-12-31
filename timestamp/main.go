// Copyright 2014 Google Inc. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

// The timestamp tool prints timestamps in various formats.
package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"willnorris.com/go/newbase60"
)

const day = 24 * time.Hour

var (
	epoch = time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)

	// inputFormats identifies the time formats used to parse user input.
	inputFormats = []string{
		time.RFC3339,
		"2006-01-02",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
	}

	//flags
	utc            = flag.Bool("utc", false, "parse times without timezones as UTC")
	printRFC3339   = flag.Bool("rfc3339", false, "print rfc 3339 timestamp only")
	printEpochDays = flag.Bool("epoch", false, "print epoch days only")
)

const usageText = `timestamp is a tool for printing timestamps in various formats.

Usage:
  timestamp [-utc] [-rfc3339] [-epoch] [time]

timestamp will print the specified time in the following formats:
  - unix timestamp (number of seconds since January 1, 1970 UTC)
  - rfc 3339 timestamp in the specified timezone (if not UTC)
  - rfc 3339 timestamp in UTC
  - ordinal date (year and day of the year) in the specified timezone
  - ordinal date (year and day of the year) in UTC (if different than above)
  - epoch days (number of days since January 1, 1970 UTC) as decimal and
    sexigesimal (newbase60) formatted. This is only printed if date is after
    1970-01-01, and is always calculated based on UTC time.

time can be specified as a full rfc 3339 timestamp, or just the date component
(YYYY-MM-DD).  If no time is specified, the current system time will be used.

time values without an explicit timezone will be interpreted as the local
system timezone unless the -utc flag is provided.

Flags:
`

func usage() {
	fmt.Fprintf(os.Stderr, usageText)
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	flag.Parse()

	loc := time.Local
	if *utc {
		loc = time.UTC
	}

	t := parseInput(flag.Arg(0), loc)
	epochDays := int(t.UTC().Sub(epoch) / day)

	if t.IsZero() {
		fmt.Fprintln(os.Stderr, "Unable to parse timestamp")
		os.Exit(1)
	}

	if *printRFC3339 {
		print(t.Format(time.RFC3339))
		return
	}

	if *printEpochDays {
		print(newbase60.EncodeInt(epochDays))
		return
	}

	fmt.Printf("%s\n\n", t)
	printTime("Unix Timestamp", "%d", t.Unix())

	if t.Location() != time.UTC {
		printTime("RFC 3339", "%s", t.Format(time.RFC3339))
	}
	printTime("RFC 3339 (UTC)", "%s", t.UTC().Format(time.RFC3339))

	if t.Location() != time.UTC {
		printTime("Ordinal Date", "%d-%d", t.Year(), t.YearDay())
	}
	printTime("Ordinal Date (UTC)", "%d-%d", t.UTC().Year(), t.UTC().YearDay())

	if epochDays > 0 {
		printTime("Epoch Days", "%d (%s)", epochDays, newbase60.EncodeInt(epochDays))
	}
}

func parseInput(s string, loc *time.Location) time.Time {
	if s == "" {
		return time.Now().In(loc)
	}

	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return time.Unix(i, 0).In(loc)
	}

	for _, f := range inputFormats {
		if t, err := time.ParseInLocation(f, s, loc); err == nil {
			return t
		}
	}

	return time.Time{}
}

func printTime(name, format string, a ...interface{}) {
	b := []interface{}{name + ":"}
	b = append(b, a...)
	fmt.Printf("%-19s "+format+"\n", b...)
}
