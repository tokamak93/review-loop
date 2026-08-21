package main

import (
	"encoding/json"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/tokamak93/review-loop/internal/metrics"
)

// printer writes a report and remembers the first write error, so the render
// functions below read as a report rather than as error plumbing.
type printer struct {
	w   io.Writer
	err error
}

func (p *printer) printf(format string, args ...any) {
	if p.err != nil {
		return
	}
	_, p.err = fmt.Fprintf(p.w, format, args...)
}

func (p *printer) flush(tw *tabwriter.Writer) {
	if p.err != nil {
		return
	}
	p.err = tw.Flush()
}

func writeJSON(out io.Writer, report metrics.Report, w window) error {
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(struct {
		Range string `json:"range"`
		metrics.Report
	}{Range: w.describe(), Report: report})
}

func writeText(out io.Writer, report metrics.Report, w window) error {
	p := &printer{w: out}
	p.printf("review-loop precision — %s\n", w.describe())
	p.printf("%d findings across %d pull requests\n\n", report.Records, report.PRs)

	writeHeadline(p, report.Overall)
	if report.Provisional {
		p.printf("\nPROVISIONAL — %s. Propose no rules from this.\n", report.Shortfall)
	}
	writeHoldout(p, report.Holdout)
	for _, dim := range report.Dimensions {
		writeDimension(p, dim)
	}
	return p.err
}

func writeHeadline(p *printer, s metrics.Summary) {
	p.printf("precision          %s   %d / %d\n", ratio(s.Precision), s.Accepted, s.Counted())
	p.printf("engaged precision  %s   %d / %d\n", ratio(s.EngagedPrecision), s.Accepted, s.Engagement())
	p.printf("  accepted %d · rejected %d · ignored %d · void %d\n",
		s.Accepted, s.Rejected, s.Ignored, s.Void)
}

func writeHoldout(p *printer, c metrics.Comparison) {
	p.printf("\nholdout — the same reviewer with the learned rules switched off\n")
	tw := tabwriter.NewWriter(p.w, 0, 0, 2, ' ', 0)
	inner := &printer{w: tw}
	inner.printf("  rules on\t%s\t%d / %d\n", ratio(c.Learning.Precision), c.Learning.Accepted, c.Learning.Counted())
	inner.printf("  rules off\t%s\t%d / %d\n", ratio(c.Holdout.Precision), c.Holdout.Accepted, c.Holdout.Counted())
	p.err = inner.err
	p.flush(tw)
	if c.Comparable {
		return
	}
	p.printf("  not comparable — one side has not cleared the floors, so any\n")
	p.printf("  improvement here is unverified. docs/metric.md, threat 4.\n")
}

func writeDimension(p *printer, dim metrics.Dimension) {
	p.printf("\nby %s\n", dim.Name)
	if dim.Degenerate {
		p.printf("  degenerate — every counted finding shares one value (%s), so this\n", soleValue(dim))
		p.printf("  segmentation measures nothing at the current configuration.\n")
		return
	}
	tw := tabwriter.NewWriter(p.w, 0, 0, 2, ' ', 0)
	inner := &printer{w: tw}
	for _, seg := range dim.Segments {
		inner.printf("  %s\t%s\t%d / %d\t%d PRs\n",
			seg.Value, ratio(seg.Summary.Precision), seg.Summary.Accepted, seg.Summary.Counted(), seg.PRs)
	}
	p.err = inner.err
	p.flush(tw)
}

func soleValue(dim metrics.Dimension) string {
	if len(dim.Segments) == 0 {
		return "none"
	}
	return dim.Segments[0].Value
}

// ratio prints a dash rather than 0.00 when there is nothing to divide by, so
// an empty bucket never reads as a measured zero.
func ratio(f func() (float64, bool)) string {
	v, ok := f()
	if !ok {
		return "   — "
	}
	return fmt.Sprintf("%5.2f", v)
}
