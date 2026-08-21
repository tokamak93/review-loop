---
name: review-harvest
description: Reads what humans did with past pr-review comments, records the outcomes, computes precision with the metrics command, and proposes changes to the learned rules file. Use on a schedule, or when asked to harvest review outcomes or update the learned rules.
---

# Harvesting review outcomes

You are closing the loop. The `pr-review` skill posted comments; humans resolved, replied to, reacted to, or ignored them. Your job is to turn that into evidence, compute the number, and propose a change to `.review/learned-rules.md`.

You do not edit the rules file directly. You open a pull request. A human approves the lesson or does not.

## The definitions are not yours to choose

`docs/metric.md` defines accepted, rejected, ignored and void, and how each is observed. Read it before you start and follow it exactly. If a case does not fit the definitions, record it as void with a reason — do not improvise a fifth category, and do not stretch an existing one to fit.

The two rules that matter most, because they are the easy ones to get wrong:

**Ignored is not rejected.** Silence means the finding was wrong, or nobody read it, or the pull request was urgent. These have opposite implications. Ignored findings are counted and reported, and are **never** used as evidence that a finding was wrong.

**Resolution alone is not acceptance.** People resolve threads to clear their inbox. Acceptance needs the anchored lines to have changed, or an explicit reply or reaction.

## Where to start

The watermark is the newest `pr_closed_at` already in `.review/outcomes.jsonl`. Read the last line; if the file is missing or empty, start from ninety days ago.

```
gh pr list --repo "$REPO" --state closed --limit 200 \
  --json number,closedAt,mergedAt,url \
  --search "closed:>=$WATERMARK"
```

Work over closed pull requests only. A finding's outcome is not settled while the branch is still moving, and re-harvesting an open one produces a record you would have to correct later — which this file's append-only rule does not allow.

Re-running over a pull request already in the history must produce nothing new. Skip any comment whose `comment_id` is already recorded; the metrics command rejects the whole file on a duplicate, so a non-idempotent run is caught rather than absorbed.

## Gathering

Find your own comments by the marker the review skill leaves:

```
gh api "repos/$REPO/pulls/$PR_NUMBER/comments" --paginate \
  --jq '.[] | select(.body | contains("<!-- review-loop"))'
```

Pull-request-level findings are issue comments, not review comments — check both endpoints, or every `pr-shape` finding silently vanishes from the history.

For each comment, parse every field out of the marker and collect: the thread's resolution state and who resolved it; replies; reactions; and whether commits after the comment's timestamp touched the anchored lines.

Thread resolution is only in the GraphQL API — `pullRequestReviewThread` — not REST. Reactions are a separate REST endpoint.

## Recording

Append one record per finding to `.review/outcomes.jsonl`, and **never rewrite an existing record**. The history is the evidence base; a rewrite destroys the ability to recompute an old number.

The schema is the `Record` type in `internal/outcome/outcome.go` in the review-loop repository. Three marker names are shorter than their field names, because a marker is written into every comment and a record is written once:

| Marker | Field |
|---|---|
| `gate=` | `confidence_gate` |
| `skill=` | `skill_sha` |
| `rules=` | `rules_sha` |

Every other marker name is the field name. Emit exactly those JSON field names — every one of them, including the marker fields `confidence_gate`, `learning`, `skill_sha` and `rules_sha`. The reader rejects unknown fields and rejects records with any required field empty, so a guess does not get absorbed quietly. If a marker is missing a field the schema requires, record the finding as `void` with that as the reason rather than inventing a value.

Commit the appended file directly to the default branch, with a message naming the pull requests harvested. This is the one thing here that does not go through review: outcomes are append-only evidence, not behaviour, a rewrite of them is visible in git history, and a weekly pull request to merge a log file is a ritual nobody would read. The learned rules are behaviour and go through a pull request. `docs/adr/0008-distribution.md` records the distinction.

## Computing

**Run the metrics command. Do not compute any figure yourself.**

```
review-metrics --file .review/outcomes.jsonl
```

Quote its output verbatim in your summary and in any pull request you open. It reports precision, engaged precision, the segmentations, the sample floors and the holdout comparison, and it produces the same numbers on the same file every time — which is the entire reason it exists and you do not. If it fails, report the failure and stop; do not fall back to arithmetic.

Three things in its output that mean something specific:

**`PROVISIONAL`** — the floors are not met. Report the numbers with the counts beside them and **propose no rules**. An early rule learned from six observations is a superstition that will suppress findings for months.

**`degenerate`** on a segmentation — every counted finding shares one value there, so the segmentation measures nothing. With the confidence gate at `high`, `by confidence` is degenerate by construction: nothing below the gate was ever posted, so nothing below it can be counted. That is expected, and it means the calibration check over confidence is unavailable rather than passing. Say so; do not report it as though confidence were validated.

**`not comparable`** on the holdout — one of the two slices has not cleared the floors. Until it does, any improvement the loop appears to have made is unverified, and `docs/metric.md` threat 4 says the README must say so.

Beyond the command's output, check calibration explicitly on the dimensions that are not degenerate: does precision rise with claimed severity? If it does not, the field is noise, and that is a finding about the review skill worth reporting even though it proposes no rule.

## What you cannot see

The review re-runs on every push and suppresses points already made, so a finding appears once per pull request even when the review ran six times. If you find the same finding posted twice on one pull request, that is a suppression bug in the review skill — record one as void with that reason and report it, rather than counting it twice.

Suppressed findings are not in your reach at all. They were never posted, so the GitHub API has no trace of them, and `concurrence` — the count of times the reviewer reached a conclusion a human had already reached — is reported by the review run in its own job summary and nowhere else. Do not attempt to reconstruct it here, and do not report a figure for it.

## Proposing rules

A rule may be proposed only from accepted and rejected outcomes. Never from ignored ones.

A proposal needs: a pattern that appears at least three times, a rule stated as a condition and an instruction, and the specific findings that motivated it, linked.

Two shapes are useful:

- **Suppression** — "stop reporting X" — from repeated rejections in a narrow, describable category
- **Sharpening** — "when you report X, check Y first" — from a mix of acceptances and rejections in one category, where the difference between them is describable

Prefer sharpening. Suppression rules are the ones that quietly cap the reviewer's ceiling, because a suppressed finding is never posted and so can never be shown to have been right. That risk is recorded in `docs/metric.md`, the holdout exists to bound it, and every suppression proposal must state it.

Propose per category, and check what else shares that category first. The categories are deliberately split so that one suppression rule cannot silence two unrelated kinds of finding — `duplication` and `antipattern` are separate for exactly this reason. A proposal that would suppress a whole category must say which agents produce it.

Keep the file bounded. If it is at its cap, the proposal must retire the least-supported existing rule and say which and why — do not simply append.

## The pull request

One pull request per harvest run, against `.review/learned-rules.md`, containing:

- The proposed rule text
- The evidence: counts, and links to the findings behind it
- The metrics command's output, verbatim, and the range it was computed over
- For any suppression rule, what it will stop reporting and what that costs if it is wrong

Title it plainly. Do not merge it, do not approve it, and do not open it if there is nothing supported to propose — a harvest run that proposes nothing is a normal outcome and by far the most common one.
