# 0007 — Confidence is computed, and it gates posting

**Status**: accepted — refines 0005

## Context

ADR 0005 rejected "a confidence threshold instead of a verdict", on the grounds that it trusts precisely the self-assessed number the ADR existed because we do not trust. That reasoning was right about the number it was talking about and does not settle the question, because two things have changed since.

First, there is now a verifier that sets confidence itself, on a fresh context, having never seen the specialist's reasoning. A threshold on *that* number is not a threshold on self-assessment.

Second, confidence had no definition. Six agents each set a field called `confidence` from whatever the word suggested to them, and the harvester then segmented precision by it as though the values were commensurable. They were not. A `high` from pr-shape and a `high` from pr-patterns were different claims about different things, averaged together into a number that meant nothing.

## Decision

**Confidence is calculated from what was read, not felt about the claim.** The rubric is `.claude/skills/pr-review/CONFIDENCE.md`, one file, referenced by every agent that sets the field. It starts at high and takes away: named demotions for not opening the file, not reading a caller the claim depends on, citing a convention not found in a document, or depending on anything outside the repository. Take the lowest cap that applies; two mediums do not make a high.

**The verifier sets it from scratch and its number is the one that counts.** The specialist's value is passed through only so a large gap is visible. `CONFIRMED` with `low` confidence is a coherent result and the verifier is told to return it when it is true.

**`REVIEW_MIN_CONFIDENCE` gates posting, defaulting to `high`.** Applied in aggregation, before ranking and before the cap, so the cap is spent on findings that were going to post.

The threshold ADR 0005 rejected was a *replacement* for the verdict. This one sits after it: a finding must be CONFIRMED **and** clear the gate. Both filters, not either.

## Consequences

**Most findings will not post.** That is the intent and it should be stated plainly to anyone adopting this, because the first week will look like the reviewer is broken.

**Confidence stops being measurable.** This is the real cost and it is the same shape as the self-reinforcement risk in threat 4. With the gate at `high`, nothing below the gate is posted, so nothing below it is counted, so the check that asks whether precision rises with claimed confidence has a single value to work with. The metrics command detects the single-valued dimension and prints `degenerate`; `gate` is recorded on every finding so the history says which configuration produced it; and `docs/metric.md` carries it as threat 5, beside threat 4, because they are the same mistake applied to two different fields.

Buying the check back means lowering the gate to `medium` for a stated window and measuring what the extra findings did. That is a real experiment someone must actually run. Until it is run, this project should not claim its confidence field is calibrated — only that it is defined.

**The pressure to inflate is now direct.** Marking a finding `high` is how it reaches a human. The rubric says so out loud rather than pretending the incentive is not there, and the verifier's independent re-scoring is the only thing that actually resists it. If the calibration check ever runs and shows `high` findings are not accepted more often than `medium` ones, the honest response is to drop the field, not to retune the rubric.

## Alternatives rejected

**Gate on the specialist's confidence.** What ADR 0005 rejected, and still wrong for the same reason.

**Gate on severity instead.** Severity is how much damage the finding does if real. Gating on it posts confident nonsense about important things, which is the worst possible failure for a reviewer whose only asset is that its comments are worth reading.

**A numeric confidence, 0–100.** More expressive, and it invites 72 as a way of avoiding a decision. Three levels with written demotions force the agent to name which demotion applied, which is checkable; a number is not.

**Leave confidence undefined and gate anyway.** The cheapest change, and it gates on a field that means six different things. Defining it was the larger half of this decision.
