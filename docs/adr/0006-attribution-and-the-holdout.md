# 0006 — Attribution, and a holdout the loop cannot see

**Status**: accepted

## Context

`docs/metric.md` promised a mitigation it did not have, and `SPEC.md` §3 required a field nobody had implemented. Both gaps point at the same hole.

The spec requires every outcome to link back to "the reviewer version that produced it". The comment marker carried category, severity, confidence, kind, agent, verdict and effort — nothing identifying the text that produced the finding. Under ADR 0003 the reviewer *is* text in git, so without that field a change in precision cannot be attributed to the change in prompt that caused it. That is the entire claim of the project, unmeasurable.

Separately, threat 4 in the metric document names a held-out fraction of pull requests reviewed with learning disabled as the mitigation for the loop's self-reinforcement risk. There was no toggle, no selection rule, and the review skill read the learned rules unconditionally. A stated mitigation with no mechanism is worse than an admitted gap, because it reads as solved.

## Decision

**Three attribution fields on every marker**, and therefore on every outcome record:

- `skill` — the commit of the repository the reviewer was installed from. The workflow passes it as `REVIEWER_VERSION` from `github.job_workflow_sha`, which is the reusable workflow's own commit; the skill falls back to `git log -1 -- .claude/` when it is running from a local checkout.
- `rules` — the revision of `.review/learned-rules.md` in force.
- `learning` — `on`, or `off` when the pull request is in the holdout.

**A holdout selected arithmetically.** `REVIEW_HOLDOUT_EVERY` defaults to 5; a pull request is held out when its number divides by it exactly. In the holdout the skill does not read the learned rules at all.

The selection rule is deliberately dumb. It is deterministic, auditable from the pull request number alone, requires no state, and — the point — cannot be influenced by the thing being measured. Anything cleverer would be a rule the reviewer could, in principle, be tuned against.

**The metrics command reports both slices** and refuses to call them comparable until each independently clears the 80-finding and 20-pull-request floors.

## Consequences

**A fifth of the loop's learning is spent on measurement.** Held-out pull requests get a reviewer that has learned nothing, which is slightly worse for those authors. That is the price of being able to say the loop works, and this project has no reason to exist if it cannot say that.

**Precision can now be recomputed per reviewer version.** `--skill <sha>` filters the history, which makes a prompt change a thing you can measure after the fact rather than a thing you argue about.

**The holdout can be switched off**, and a repository that switches it off cannot report a verified improvement. The README must then say the improvement is unverified rather than dropping the qualifier — recorded here so that dropping it later is visibly a decision.

**It does not fix threat 6.** A finding the verifier wrongly rejected is invisible to the holdout too: it was never posted on either slice. The holdout bounds the damage a *learned rule* can hide, and nothing else.

## Alternatives rejected

**Random holdout selection.** Statistically cleaner, and it makes a run irreproducible and an outcome record impossible to explain after the fact. The arithmetic rule can be checked against the pull request number by hand.

**A holdout of alternating weeks.** Confounds the slice with whatever else happened that week — a release, a migration, one author on holiday. Interleaving by pull request number spreads both slices across the same conditions.

**Recording the skill version only in the outcome record, not the marker.** Cheaper, and it loses the version for any comment harvested after the file changed. The marker is written at post time and never moves, which is the only place the fact is reliably true.
