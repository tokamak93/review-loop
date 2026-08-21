# Calculating confidence

Every agent in this pipeline sets a confidence on every finding, and every agent uses this file to set it. One rubric, one meaning, because the number is compared across agents later and a number that means something different per agent compares nothing.

**Confidence is not how the claim feels. It is a statement about what you read.** Severity is how much damage the finding does; confidence is how much of the claim you verified with your own eyes. They move independently — a catastrophic bug you inferred is high severity and low confidence, and it does not get posted.

## The procedure

Start at **high** and take away. You do not argue your way up.

**high** — every input the claim depends on is in this repository, and you opened and read all of it. You can walk the failure scenario through specific lines you have seen. You looked for the thing that would make the claim wrong and it is not there.

**medium** — you read the code and the claim holds against it, but one input is inferred rather than read.

**low** — you are reasoning from a pattern rather than from this code.

## The demotions

These apply regardless of how convincing the claim is. Each caps confidence at the level shown.

| If | At most |
|---|---|
| You read the diff hunk but did not open the file around it | medium |
| The claim depends on how a function is called and you did not read a caller | medium |
| You cite a convention you did not find written in a document you read | medium |
| The behaviour depends on configuration, a runtime value, or a service outside this repository | medium |
| You could not walk the scenario through concrete lines | low |
| You searched for a contradicting guard and did not have time to finish | low |

Take the lowest cap that applies. Two mediums do not make a high.

## The pressure, stated plainly

The orchestrator posts only findings at or above `$REVIEW_MIN_CONFIDENCE`, which defaults to **high**. So marking a finding `high` is how it reaches a human, and marking it honestly is how it does not.

That pressure is real and there is no point pretending otherwise. Three things exist because of it:

1. **pr-verify sets confidence again, from scratch**, having never seen yours. Its number is the one that decides. Inflating yours does not survive the pass; it just wastes it.
2. **Confidence is recorded in the comment marker** and precision is segmented by it. A specialist whose `high` findings are not accepted more often than its `medium` ones has a confidence field that means nothing, and `docs/metric.md` says to drop the field rather than keep pretending.
3. **The correct output is often nothing.** An agent that returns no findings has not failed. An agent that returns four inflated ones has.

## What this costs

With the gate at high, no medium-confidence finding is ever posted, so none is ever counted, so the check in point 2 above has nothing to compare against — the `by confidence` segmentation goes degenerate and the metrics command says so out loud rather than printing one row that looks like a result.

That is a real loss and it is the price of the gate. The way to buy it back is to lower the gate deliberately for a stated window and measure what the extra findings did. `docs/adr/0007-confidence-is-computed-and-gated.md` records this; do not quietly leave the gate at high forever and then claim confidence is calibrated.
