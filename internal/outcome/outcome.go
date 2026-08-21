// Package outcome defines the record the review-harvest skill appends to
// .review/outcomes.jsonl, and reads it back.
//
// The record type in this file is the schema. There is no second copy of it in
// prose: docs/metric.md points here, and the harvest skill is told to emit
// exactly these fields. A record that does not round-trip through Load is a
// harvester bug, and Load says so rather than skipping the line.
package outcome

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// State is the outcome of a single posted finding, as defined in
// docs/metric.md. Exactly one applies, and it is decided after the pull
// request closes.
type State string

// The four outcome states. Void is excluded from every denominator; the other
// three make up the precision denominator.
const (
	Accepted State = "accepted"
	Rejected State = "rejected"
	Ignored  State = "ignored"
	Void     State = "void"
)

// Valid reports whether s is one of the four defined states.
func (s State) Valid() bool {
	switch s {
	case Accepted, Rejected, Ignored, Void:
		return true
	default:
		return false
	}
}

// Category is the kind of defect a finding describes, from the closed list in
// the pr-review skill's finding contract.
//
// It is a closed type for the same reason State is. Precision is segmented by
// category and learned rules are proposed per category, so an invented value
// does not merely look untidy — it silently splits one bucket in two and
// halves both counts in exactly the figure a rule would be proposed from. A
// specialist that invents a category is a bug, and Load says so.
type Category string

// Every category a finding may carry, grouped by the agent that produces it.
const (
	Correctness   Category = "correctness"
	Concurrency   Category = "concurrency"
	ErrorHandling Category = "error-handling"
	ResourceLeak  Category = "resource-leak"

	Antipattern   Category = "antipattern"
	APIContract   Category = "api-contract"
	DataChange    Category = "data-change"
	Observability Category = "observability"

	Duplication Category = "duplication"
	Efficiency  Category = "efficiency"

	TestGap  Category = "test-gap"
	Security Category = "security"

	PRShape       Category = "pr-shape"
	ScopeMismatch Category = "scope-mismatch"
)

var categories = map[Category]struct{}{
	Correctness: {}, Concurrency: {}, ErrorHandling: {}, ResourceLeak: {},
	Antipattern: {}, APIContract: {}, DataChange: {}, Observability: {},
	Duplication: {}, Efficiency: {},
	TestGap: {}, Security: {},
	PRShape: {}, ScopeMismatch: {},
}

// Valid reports whether c is one of the defined categories.
func (c Category) Valid() bool {
	_, ok := categories[c]
	return ok
}

// Categories returns every defined category, for callers that need to present
// the closed list rather than restate it.
func Categories() []Category {
	out := make([]Category, 0, len(categories))
	for c := range categories {
		out = append(out, c)
	}
	return out
}

// Record is one posted finding and what became of it.
//
// The fields above Outcome are copied from the finding's comment marker and
// are fixed at post time; the fields below it are the harvester's judgement,
// made after the pull request closed.
type Record struct {
	// Identity.
	FindingID  string `json:"finding_id"`
	Repo       string `json:"repo"`
	PR         int    `json:"pr"`
	CommentID  int64  `json:"comment_id"`
	CommentURL string `json:"comment_url"`

	// Location. Kind is "inline" or "pr-level"; Path and Line are empty for
	// pr-level findings.
	Kind string `json:"kind"`
	Path string `json:"path,omitempty"`
	Line int    `json:"line,omitempty"`

	// Claimed at post time, from the marker.
	Category   Category `json:"category"`
	Severity   string   `json:"severity"`
	Confidence string   `json:"confidence"`
	Agent      string   `json:"agent"`
	Verdict    string   `json:"verdict"`
	Effort     string   `json:"effort"`

	// Gate is the minimum confidence that was required to post at the time.
	// It is recorded because it decides what is absent from the history: with
	// a gate of "high", no medium-confidence finding exists to be counted, and
	// any calibration check over confidence is degenerate rather than clean.
	Gate string `json:"confidence_gate"`

	// Attribution. Learning is "on" or "off": "off" means this pull request
	// was in the holdout and the reviewer did not read the learned rules.
	// SkillSHA and RulesSHA identify the reviewer version that produced the
	// finding, so a change in precision can be attributed to a change in text.
	Learning string `json:"learning"`
	SkillSHA string `json:"skill_sha"`
	RulesSHA string `json:"rules_sha"`

	// Timing.
	PostedAt time.Time `json:"posted_at"`
	ClosedAt time.Time `json:"pr_closed_at"`
	Merged   bool      `json:"pr_merged"`

	// The harvester's judgement. Reason must name the signal it read.
	Outcome     State     `json:"outcome"`
	Reason      string    `json:"outcome_reason"`
	HarvestedAt time.Time `json:"harvested_at"`
}

// InHoldout reports whether the finding was produced with the learned rules
// switched off.
func (r Record) InHoldout() bool { return r.Learning == "off" }

// PathPrefix returns the leading one or two path segments, the unit
// docs/metric.md segments precision by. It is empty for pr-level findings.
func (r Record) PathPrefix() string {
	if r.Path == "" {
		return ""
	}
	var start, seen int
	for i := range len(r.Path) {
		if r.Path[i] != '/' {
			continue
		}
		seen++
		if seen == 2 {
			return r.Path[:i]
		}
		start = i
	}
	if start == 0 {
		return ""
	}
	return r.Path[:start]
}

// Validate reports the first reason the record cannot be counted. Missing
// fields are errors rather than defaults: a record silently counted with an
// empty category segments into a bucket that means nothing.
func (r Record) Validate() error {
	if err := r.validateIdentity(); err != nil {
		return err
	}
	if !r.Category.Valid() {
		return fmt.Errorf("category is %q, which is not one of the defined categories", r.Category)
	}
	for _, f := range []struct{ name, value string }{
		{"severity", r.Severity},
		{"confidence", r.Confidence},
		{"agent", r.Agent},
		{"verdict", r.Verdict},
		{"effort", r.Effort},
		{"confidence_gate", r.Gate},
		{"skill_sha", r.SkillSHA},
	} {
		if f.value == "" {
			return fmt.Errorf("%s is empty", f.name)
		}
	}
	if r.Learning != "on" && r.Learning != "off" {
		return fmt.Errorf("learning is %q, want on or off", r.Learning)
	}
	if !r.Outcome.Valid() {
		return fmt.Errorf("outcome is %q, want accepted, rejected, ignored or void", r.Outcome)
	}
	if r.Outcome == Void && r.Reason == "" {
		return fmt.Errorf("void outcome with no reason")
	}
	return nil
}

func (r Record) validateIdentity() error {
	if r.FindingID == "" {
		return fmt.Errorf("finding_id is empty")
	}
	if r.Repo == "" {
		return fmt.Errorf("repo is empty")
	}
	if r.PR <= 0 {
		return fmt.Errorf("pr is %d, want a positive number", r.PR)
	}
	if r.CommentID <= 0 {
		return fmt.Errorf("comment_id is %d, want a positive number", r.CommentID)
	}
	return nil
}

// Load reads newline-delimited records, rejecting the file on the first
// invalid or duplicated one.
//
// Duplicate comment IDs are an error rather than a deduplication: the
// harvester is required to be idempotent, and a repeated comment means it is
// not. Silently collapsing the duplicate would hide the bug and leave the
// counts wrong in a way nobody would notice.
func Load(r io.Reader) ([]Record, error) {
	var (
		records []Record
		seen    = map[int64]int{}
		line    int
	)
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		line++
		text := scanner.Bytes()
		if len(text) == 0 {
			continue
		}
		rec, err := decode(text, seen, line)
		if err != nil {
			return nil, err
		}
		seen[rec.CommentID] = line
		records = append(records, rec)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading outcomes: %w", err)
	}
	return records, nil
}

func decode(text []byte, seen map[int64]int, line int) (Record, error) {
	var rec Record
	dec := json.NewDecoder(bytes.NewReader(text))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&rec); err != nil {
		return Record{}, fmt.Errorf("line %d: %w", line, err)
	}
	if err := rec.Validate(); err != nil {
		return Record{}, fmt.Errorf("line %d: %w", line, err)
	}
	if first, dup := seen[rec.CommentID]; dup {
		return Record{}, fmt.Errorf("line %d: comment %d already recorded on line %d", line, rec.CommentID, first)
	}
	return rec, nil
}
