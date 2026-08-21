# Learned rules

What this team has confirmed by acting on, or dismissing, findings from the `pr-review` skill. The skill reads this file before every review and it overrides the skill's defaults.

**This file is not edited by hand and not written by the action.** The `review-harvest` skill proposes changes as a pull request, with the evidence attached, and a human approves or rejects them. That is the only path in.

Empty until there is evidence. A rule here before the thresholds in `docs/metric.md` are met is a guess wearing the costume of a finding.

## Format

Each rule:

```
### <identifier> — <one-line summary>

**Rule**: <condition, then instruction>

**Evidence**: <counts> across <n> pull requests, <date range>
<links to the findings>

**Added**: <date> · **Kind**: sharpening | suppression
```

Suppression rules also carry:

```
**Cost if wrong**: <what stops being reported, and how that would be noticed>
```

## Cap

Twelve rules. At the cap, a proposal must retire the least-supported existing rule and say which and why. An unbounded rules file becomes a prompt nobody reads, including the model.

## Rules

_None yet._
