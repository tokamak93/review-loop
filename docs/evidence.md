# Evidence

**Nothing has been measured yet.** No run of this reviewer has happened against any pull request, real or otherwise. Every number in this file is absent rather than pending.

This file exists in this state on purpose. It is the deliverable of SPEC §6, it is where the project's central claim either survives or does not, and leaving it uncreated until there was something flattering to put in it is how a repository ends up with a README full of assertions and no page to check them against.

## What will go here

| | |
|---|---|
| **The repository** | Named, linked. Not a synthetic corpus. |
| **The pull requests** | Linked, individually. |
| **Baseline precision** | Loop off, over at least 80 non-void findings across at least 20 pull requests |
| **Precision with the loop on** | Same floors, same command, stated range |
| **The holdout comparison** | `learning=on` against `learning=off`, and whether the metrics command called it comparable |
| **Accepted findings** | Examples, quoted |
| **Rejected findings** | Examples, quoted, in the same number and the same detail |

Every figure comes from `go run ./cmd/metrics`, with the range it was computed over printed beside it. No figure is typed in by hand, including the ones that look obvious.

## What this file must say if the claim fails

If precision does not improve, that is the result, and it is reported here in the same detail a success would get — same sections, same examples, same range.

If the holdout has not cleared the floors, the improvement is reported as **unverified**, in that word, and the README carries the same qualifier. `docs/failure-modes.md` explains why: a learned rule that suppresses a category raises precision whether or not the rule was any good, and without the holdout there is no way to tell those apart.

If the reviewer turns out to be worse than posting nothing, that goes here too.

## Status

Nothing run. `.review/outcomes.jsonl` is empty. The first thing to do is measure the baseline before any learned rule exists — SPEC's suggested order, step 4 — because once a rule is in the file there is no way back to an unmodified reviewer over the same pull requests.
