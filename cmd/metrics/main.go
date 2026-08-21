// Command metrics computes the precision figures defined in docs/metric.md
// over the harvested outcome history.
//
// It reads only committed data and calls no APIs, so the same file produces
// the same numbers every time. Every figure quoted anywhere in this repository
// comes from this command, with the range it was computed over next to it.
//
//	go run ./cmd/metrics
//	go run ./cmd/metrics --since 2026-01-01 --skill 4f21ac9
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/tokamak93/review-loop/internal/metrics"
	"github.com/tokamak93/review-loop/internal/outcome"
)

type options struct {
	file  string
	since string
	until string
	skill string
	json  bool
}

func main() {
	var opts options
	flag.StringVar(&opts.file, "file", ".review/outcomes.jsonl", "outcome history to read")
	flag.StringVar(&opts.since, "since", "", "only findings posted on or after this date (YYYY-MM-DD)")
	flag.StringVar(&opts.until, "until", "", "only findings posted before this date (YYYY-MM-DD)")
	flag.StringVar(&opts.skill, "skill", "", "only findings produced by this reviewer version (commit sha prefix)")
	flag.BoolVar(&opts.json, "json", false, "emit the report as JSON")
	flag.Parse()

	if err := run(opts, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "metrics:", err)
		os.Exit(1)
	}
}

func run(opts options, out io.Writer) error {
	f, err := os.Open(opts.file)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	records, err := outcome.Load(f)
	if err != nil {
		return err
	}
	window, err := parseWindow(opts)
	if err != nil {
		return err
	}
	records = window.filter(records)
	if len(records) == 0 {
		return errors.New("no records in range")
	}
	report := metrics.Compute(records)
	if opts.json {
		return writeJSON(out, report, window)
	}
	return writeText(out, report, window)
}

type window struct {
	since time.Time
	until time.Time
	skill string
}

func parseWindow(opts options) (window, error) {
	w := window{skill: opts.skill}
	var err error
	if opts.since != "" {
		if w.since, err = time.Parse(time.DateOnly, opts.since); err != nil {
			return window{}, fmt.Errorf("--since: %w", err)
		}
	}
	if opts.until != "" {
		if w.until, err = time.Parse(time.DateOnly, opts.until); err != nil {
			return window{}, fmt.Errorf("--until: %w", err)
		}
	}
	return w, nil
}

func (w window) filter(records []outcome.Record) []outcome.Record {
	kept := make([]outcome.Record, 0, len(records))
	for _, r := range records {
		if !w.since.IsZero() && r.PostedAt.Before(w.since) {
			continue
		}
		if !w.until.IsZero() && !r.PostedAt.Before(w.until) {
			continue
		}
		if w.skill != "" && !hasPrefix(r.SkillSHA, w.skill) {
			continue
		}
		kept = append(kept, r)
	}
	return kept
}

func (w window) describe() string {
	span := "all recorded outcomes"
	switch {
	case !w.since.IsZero() && !w.until.IsZero():
		span = fmt.Sprintf("%s to %s", w.since.Format(time.DateOnly), w.until.Format(time.DateOnly))
	case !w.since.IsZero():
		span = "since " + w.since.Format(time.DateOnly)
	case !w.until.IsZero():
		span = "until " + w.until.Format(time.DateOnly)
	}
	if w.skill != "" {
		span += ", reviewer " + w.skill
	}
	return span
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
