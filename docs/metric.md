# The metric

Written before the harvester, before the reviewer, and before any learning. Everything this repository claims is measured against the definitions below, so they are fixed here and changed only by amending this document.

## What is being measured

**Precision.** Of the findings posted, what fraction did a human act on.

Not recall. This repository cannot measure recall, because the defects the reviewer failed to mention are unobservable — there is no ground truth for "bugs present in this pull request". Any claim about recall here would be invented. The README says so, and no chart in this repository has recall on an axis.

Precision is the right target anyway. A reviewer at 30% precision trains the team to skim past every comment it writes, and at that point its recall is irrelevant because nobody is reading it.

## Outcome classification

Every posted finding ends in exactly one of four states. The classifier runs after the pull request closes, never before, because a finding's state is not settled while the branch is still moving.

### Accepted

Any of:

- The comment thread was resolved by someone other than the pull request author, and the lines it pointed at changed after the comment was posted
- The comment thread was resolved by the author and the lines changed after the comment was posted
- A commit after the comment modified the lines the comment anchored to, and the change is not a pure rebase or formatting pass
- A human replied agreeing, or reacted with 👍 / 🎯

The line-change signal is the load-bearing one. Resolution alone is weak — resolving is how people clear their inbox.

### Rejected

Any of:

- A human replied disagreeing
- Reacted with 👎
- The thread was resolved with no change to the anchored lines, and the pull request merged

Explicit rejection is rare and valuable. It is the strongest training signal available and is weighted accordingly in the rules proposals.

### Ignored

The pull request merged with the thread unresolved, unreplied, unreacted, and the anchored lines unchanged.

**This is not the same as rejected, and conflating the two is the biggest threat to this metric.** Ignored means one of: the finding was wrong, or nobody read it, or the pull request was urgent, or the reviewer posted twelve findings and the team gave up at four. These have opposite implications and the signal cannot distinguish them.

Ignored findings are therefore counted separately and reported separately. They are never fed to the rules proposer as evidence that a finding was wrong. A rule may only be proposed from accepted and rejected outcomes.

### Void

Excluded from the denominator:

- The pull request was closed without merging
- The finding anchored to a line that a later force-push destroyed, so the outcome is unobservable
- The pull request was authored by the action itself — the rules proposals the harvester opens
- Fewer than two human participants on the pull request — a solo merge with no reviewer is not evidence about anything
- The comment's marker is missing a field the record schema requires, so the finding cannot be segmented

Void findings are logged with their reason. A rising void rate means the harvester is degrading and is itself worth watching.

**A voided finding stops counting as having been said.** The review skill suppresses a finding whose point is already on the pull request, and a finding voided for lost anchoring is excluded from that check. Otherwise a force-push would both void the original and silence its replacement, and the point would disappear from the pull request entirely — counted nowhere, said nowhere. The skill implements this by skipping review comments whose `position` is `null`.

### Suppressed

The review re-runs on every push, and it drops any finding whose point was already made — by an earlier run, by a human reviewer, or by the author. A suppressed finding is never posted, so it never enters the denominator. **A finding is counted once per pull request, never once per run**, and a pull request pushed to six times does not get six chances to be wrong.

Suppression is observable only in the review run's own job summary, which the workflow surfaces with `display_report: true`. The harvester cannot see it: nothing was posted, so the GitHub API has no trace. Three counts are reported there and nowhere else:

- **Self-repeat** — this reviewer already said it. Pure noise avoided; no signal beyond that.
- **Human already said it** — a human reviewer independently raised the same point. Reported as **concurrence**: the reviewer reached a conclusion a human also reached, without spending any of the team's attention to do it. It is not precision — nobody acted on a comment that was never posted — and it must never be added to the numerator or the denominator.
- **Author pre-empted it** — the author's own note says they know. Neutral.

A rising concurrence count alongside flat precision is worth investigating rather than celebrating: it can mean the reviewer is slow rather than agreeable, arriving after the human on findings it could have posted first.

## The number

```
precision = accepted / (accepted + rejected + ignored)
```

Ignored sits in the denominator deliberately. A reviewer that posts a finding nobody engages with has cost the team attention and returned nothing, and a metric that excused that would be measuring the wrong thing.

A second figure is reported alongside it and never instead of it:

```
engaged precision = accepted / (accepted + rejected)
```

This is precision among findings a human demonstrably read. It will be much higher. It is useful for diagnosis — a large gap between the two means the reviewer is posting too much, not that it is posting wrongly — and it is not the headline number, because quoting it as the headline is the obvious way to make this project look better than it is.

## Sample size

No comparison between two configurations is reported below **80 non-void findings** on each side, across at least **20 pull requests**.

The pull request floor matters as much as the finding floor. Eighty findings drawn from four pull requests measures four authors on four days, not a reviewer.

Below those floors, results are recorded but reported as provisional, with the counts shown next to them. Above them, a difference smaller than the run-to-run spread of the same configuration is not a difference.

The floors are constants in `internal/metrics/metrics.go`, so the tool refuses to present an under-powered result as a finished one rather than relying on whoever is reading to remember.

## Attribution

A number that moved is worthless without knowing what moved it. Every posted comment carries, in its marker and therefore in its outcome record:

| Field | What it lets you ask |
|---|---|
| `skill` | Which version of the reviewer's text produced this — the commit of the repository the skill was installed from |
| `rules` | Which revision of `.review/learned-rules.md` was in force |
| `learning` | Whether the learned rules were read at all, or this pull request was in the holdout |
| `gate` | What confidence floor was required to post, and therefore what is absent from the history |
| `agent` | Which specialist produced it, which validates the model assignment rather than assuming it |
| `verdict` | CONFIRMED or PLAUSIBLE, so the value of posting unverifiable findings can be computed separately |
| `effort` | Which dial position was in force |

Without `skill` and `rules`, a change in precision cannot be attributed to the change in text that caused it, and the loop cannot demonstrate anything — which is the only thing this project exists to do.

## The holdout

Every Nth pull request — five by default, `REVIEW_HOLDOUT_EVERY` — is reviewed with the learned rules switched off, and its findings are recorded `learning=off`.

This is the mitigation for threat 4 below and it is the only mechanism in the design that can catch a learned rule that improved the metric without improving the reviewer. It costs a fifth of the loop's learning.

The metrics command reports both slices and marks the comparison **not comparable** until each independently clears the floors. Until then, an improvement is a difference between two small numbers.

Setting `REVIEW_HOLDOUT_EVERY: 0` disables it. A repository that does so cannot report a verified improvement, and the README must say the improvement is unverified rather than quietly dropping the qualifier.

## Segmentation

The headline number hides the thing most worth knowing, so every report also breaks precision down by:

- **Finding category** — a reviewer at 70% on concurrency and 20% on duplication is two reviewers, and only one of them should keep running
- **Severity as claimed at post time** — if precision does not rise with claimed severity, the reviewer's self-assessment is noise and severity should be dropped from the output schema
- **Confidence as claimed at post time** — the same test, and this one directly validates whether confidence is usable as a filter
- **File path prefix** — precision on `internal/` versus `cmd/` versus generated code
- **Agent, verdict, effort and learning** — see Attribution above

The severity and confidence checks are calibration checks. They are the cheapest early evidence that the reviewer's structured output means anything at all.

**The confidence check is degenerate while the gate is set to `high`.** Nothing below the gate is ever posted, so nothing below it can ever be counted, and a segmentation with one value measures nothing. The metrics command detects this and prints `degenerate` rather than a single row that reads like a result. Buying the check back means lowering the gate deliberately for a stated window and measuring what the extra findings did; `docs/adr/0007-confidence-is-computed-and-gated.md` records the trade. Do not leave the gate at `high` indefinitely and then claim confidence is calibrated.

## How it is observed

Everything above is derived from the GitHub API, from data that exists whether or not this reviewer is installed:

| Signal | Source |
|---|---|
| Thread resolution and who resolved it | `pullRequestReviewThread` via GraphQL — REST does not expose resolution |
| Replies in a thread | Review comments filtered by `in_reply_to_id` |
| Reactions | Reactions API on the comment |
| Line changes after the comment | Commits after `comment.created_at`, diffed against the comment's `path` and `line` |
| Lost anchoring | `position: null` on the review comment |
| Merge state and participants | Pull request object |
| Suppression counts | The review run's job summary, and nowhere else |

Comments are posted individually, one per finding, for exactly this reason. A single summary comment carrying twelve findings has one resolution state and one reaction set, and every distinction in this document collapses.

## Threats to validity

Recorded here rather than in the README, so that they are impossible to quietly drop later.

1. **Silence is ambiguous.** Handled by separating ignored from rejected and never learning from ignored. It remains the weakest joint in the design.
2. **Line-change attribution is approximate.** A commit that touches the anchored lines may be unrelated to the comment. Rebases and formatting passes are filtered; coincidence is not.
3. **Politeness inflates acceptance.** People resolve threads to be agreeable. This is why resolution alone does not count as accepted.
4. **The reviewer influences its own denominator.** Once a learned rule suppresses a category, findings in that category are never posted, so they can never be rejected — and precision rises whether or not the rule was correct. The holdout above is the mitigation: a slice reviewed with learning off, recorded `learning=off`, compared by the metrics command, and reported as not comparable until both sides clear the floors. Where the holdout is disabled or has not cleared the floors, any reported improvement is unverified and must be labelled so.
5. **The confidence gate does the same thing to confidence.** With the gate at `high` the reviewer only ever posts what it claims to be sure of, which makes its confidence look perfectly calibrated by construction. Recorded as `gate` on every finding and reported as degenerate rather than as a pass. Same shape as threat 4, and it belongs beside it.
6. **A false REJECTED is silent.** A correct finding the verifier talked itself out of is never posted and never measured. Bounded only by requiring a rejection to name the specific guard, caller, test or invariant that makes the finding wrong.
7. **Small teams.** On a repository with two contributors, precision measures those two people's habits. That is fine — it is the point — but it does not generalise, and the README must not imply it does.

## Reproducibility

The metric is computed by a command anyone can run over the stored outcome history:

```
go run ./cmd/metrics --file .review/outcomes.jsonl
go run ./cmd/metrics --since 2026-01-01 --skill 4f21ac9
```

It reads only committed data, calls no APIs, and produces the same numbers on the same inputs. Every figure quoted anywhere in this repository — README, evidence document, harvest pull request summaries — comes from that command, with the range it was computed over stated next to it.

The harvest workflow installs it as `review-metrics` and the harvest skill is instructed to run it and quote it verbatim, because an agent asked to divide four hundred records by hand gives a different answer each time it is asked. The record schema it reads is the `Record` type in `internal/outcome/outcome.go`; that type is the schema, and this document does not keep a second copy of it to drift against.
