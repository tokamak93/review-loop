// Package metrics computes the precision figures defined in docs/metric.md
// from a set of harvested outcome records.
//
// It is deterministic and calls nothing: the same records produce the same
// numbers on any machine, which is the whole reason this is code and not a
// paragraph asking a model to add things up.
package metrics

import (
	"fmt"
	"sort"

	"github.com/tokamak93/review-loop/internal/outcome"
)

// The sample floors from docs/metric.md. Below either of them a report is
// provisional and no rule may be proposed from it.
const (
	MinFindings = 80
	MinPRs      = 20
)

// The dimensions docs/metric.md requires every report to segment by.
const (
	ByCategory   = "category"
	BySeverity   = "severity"
	ByConfidence = "confidence"
	ByPath       = "path"
	ByAgent      = "agent"
	ByVerdict    = "verdict"
	ByEffort     = "effort"
	ByLearning   = "learning"
)

// Dimensions is every segmentation, in report order.
var Dimensions = []string{ByCategory, BySeverity, ByConfidence, ByPath, ByAgent, ByVerdict, ByEffort, ByLearning}

// The ordinal scale shared by severity, confidence and effort.
const (
	high   = "high"
	medium = "medium"
	low    = "low"
)

// Summary counts findings by outcome.
type Summary struct {
	Accepted int
	Rejected int
	Ignored  int
	Void     int
}

func (s *Summary) add(state outcome.State) {
	switch state {
	case outcome.Accepted:
		s.Accepted++
	case outcome.Rejected:
		s.Rejected++
	case outcome.Ignored:
		s.Ignored++
	case outcome.Void:
		s.Void++
	}
}

// Counted is the precision denominator: every non-void finding.
func (s Summary) Counted() int { return s.Accepted + s.Rejected + s.Ignored }

// Engagement is the engaged-precision denominator: findings a human
// demonstrably read.
func (s Summary) Engagement() int { return s.Accepted + s.Rejected }

// Precision is accepted over every non-void finding. The second return is
// false when nothing has been counted, so callers print a dash rather than a
// zero that looks like a measurement.
func (s Summary) Precision() (float64, bool) {
	if s.Counted() == 0 {
		return 0, false
	}
	return float64(s.Accepted) / float64(s.Counted()), true
}

// EngagedPrecision is accepted over findings a human replied to, reacted to or
// resolved. It is always reported alongside Precision and never instead of it.
func (s Summary) EngagedPrecision() (float64, bool) {
	if s.Engagement() == 0 {
		return 0, false
	}
	return float64(s.Accepted) / float64(s.Engagement()), true
}

// Segment is one value of one dimension.
type Segment struct {
	Value   string
	Summary Summary
	PRs     int
}

// Dimension is a full segmentation, in descending order of sample size.
type Dimension struct {
	Name     string
	Segments []Segment
	// Degenerate marks a dimension whose records all share one value, which
	// makes it unmeasurable rather than uniform. Confidence goes degenerate as
	// soon as the posting gate is set to high: nothing below the gate is ever
	// posted, so nothing below the gate can ever be counted.
	Degenerate bool
}

// Comparison holds the holdout contrast: findings produced with the learned
// rules on, against those produced with them off.
type Comparison struct {
	Learning Summary
	Holdout  Summary
	// Comparable is false until both sides clear the floors. Until then the
	// difference between them is not evidence of anything.
	Comparable bool
}

// Report is everything docs/metric.md requires a set of outcomes to yield.
type Report struct {
	Records     int
	PRs         int
	Overall     Summary
	Dimensions  []Dimension
	Holdout     Comparison
	Provisional bool
	Shortfall   string
}

// Compute builds the report. Records are expected to have been validated by
// outcome.Load.
func Compute(records []outcome.Record) Report {
	rep := Report{Records: len(records)}
	prs := map[string]struct{}{}
	for _, r := range records {
		rep.Overall.add(r.Outcome)
		prs[prKey(r)] = struct{}{}
		if r.InHoldout() {
			rep.Holdout.Holdout.add(r.Outcome)
			continue
		}
		rep.Holdout.Learning.add(r.Outcome)
	}
	rep.PRs = len(prs)
	for _, name := range Dimensions {
		rep.Dimensions = append(rep.Dimensions, segment(records, name))
	}
	rep.Holdout.Comparable = clears(rep.Holdout.Learning, records, true) &&
		clears(rep.Holdout.Holdout, records, false)
	rep.Provisional, rep.Shortfall = floors(rep.Overall, rep.PRs)
	return rep
}

func floors(s Summary, prs int) (provisional bool, shortfall string) {
	switch {
	case s.Counted() < MinFindings && prs < MinPRs:
		return true, fmt.Sprintf("%d of %d findings and %d of %d pull requests",
			s.Counted(), MinFindings, prs, MinPRs)
	case s.Counted() < MinFindings:
		return true, fmt.Sprintf("%d of %d findings", s.Counted(), MinFindings)
	case prs < MinPRs:
		return true, fmt.Sprintf("%d of %d pull requests", prs, MinPRs)
	default:
		return false, ""
	}
}

func clears(s Summary, records []outcome.Record, learning bool) bool {
	if s.Counted() < MinFindings {
		return false
	}
	prs := map[string]struct{}{}
	for _, r := range records {
		if r.InHoldout() == learning {
			continue
		}
		prs[prKey(r)] = struct{}{}
	}
	return len(prs) >= MinPRs
}

func segment(records []outcome.Record, name string) Dimension {
	summaries := map[string]*Summary{}
	prs := map[string]map[string]struct{}{}
	for _, r := range records {
		value := dimensionValue(r, name)
		if value == "" {
			continue
		}
		if _, ok := summaries[value]; !ok {
			summaries[value] = &Summary{}
			prs[value] = map[string]struct{}{}
		}
		summaries[value].add(r.Outcome)
		prs[value][prKey(r)] = struct{}{}
	}
	dim := Dimension{Name: name, Degenerate: len(summaries) < 2}
	for value, s := range summaries {
		dim.Segments = append(dim.Segments, Segment{Value: value, Summary: *s, PRs: len(prs[value])})
	}
	sortSegments(name, dim.Segments)
	return dim
}

// ordered dimensions read as a scale, and the calibration checks in
// docs/metric.md only mean something if they are printed in scale order.
var ordinal = map[string][]string{
	BySeverity:   {high, medium, low},
	ByConfidence: {high, medium, low},
	ByEffort:     {high, medium, low},
}

func sortSegments(name string, segments []Segment) {
	if scale, ok := ordinal[name]; ok {
		rank := map[string]int{}
		for i, v := range scale {
			rank[v] = i
		}
		sort.Slice(segments, func(i, j int) bool {
			return rank[segments[i].Value] < rank[segments[j].Value]
		})
		return
	}
	sort.Slice(segments, func(i, j int) bool {
		if segments[i].Summary.Counted() != segments[j].Summary.Counted() {
			return segments[i].Summary.Counted() > segments[j].Summary.Counted()
		}
		return segments[i].Value < segments[j].Value
	})
}

func dimensionValue(r outcome.Record, name string) string {
	switch name {
	case ByCategory:
		return r.Category
	case BySeverity:
		return r.Severity
	case ByConfidence:
		return r.Confidence
	case ByPath:
		return r.PathPrefix()
	case ByAgent:
		return r.Agent
	case ByVerdict:
		return r.Verdict
	case ByEffort:
		return r.Effort
	case ByLearning:
		return r.Learning
	default:
		return ""
	}
}

func prKey(r outcome.Record) string { return fmt.Sprintf("%s#%d", r.Repo, r.PR) }
