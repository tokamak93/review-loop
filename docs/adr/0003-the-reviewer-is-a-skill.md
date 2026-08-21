# 0003 — The reviewer is a skill, not a program

**Status**: accepted — supersedes 0001 and 0002

## Context

The first design was a compiled action: fetch the diff, assemble bounded context, make one structured-output call, post comments. ADRs 0001 and 0002 argued the context window and the call shape on that basis.

That design makes the reviewer's behaviour live in code, and the loop's whole purpose is to rewrite the reviewer's behaviour. A learning loop whose output is a patch to a TypeScript or Go source file is proposing code changes to itself, which nobody will approve at the cadence the loop needs.

The behaviour needs to live in text that a human can read in a pull request and judge in a minute.

## Decision

The reviewer is a Claude Code **skill** — `.claude/skills/pr-review/SKILL.md` — invoked by `anthropics/claude-code-action` in a workflow. The harvest half is a second skill on a schedule.

There is no compiled artifact **in the reviewer**. The workflow is YAML, the review is Markdown, and the learned rules are Markdown the harvest skill proposes changes to.

There is one in the *measurement*: `cmd/metrics` is a small Go program that reads the outcome history and divides. It was added because `docs/metric.md` promises figures anyone can recompute, and an agent asked to add up four hundred records gives a different answer each time. It never runs in the review job, never sees a diff and never calls a model — which is precisely why it can make that promise. The rule this ADR is defending is that *the reviewer's behaviour* lives in text a human can judge in a pull request, and a calculator does not touch it.

Static checks run first as a separate job, and the review job depends on them. Build, vet, lint and tests must pass before Claude is invoked.

## Consequences

**ADR 0001 is superseded.** It rejected agentic context selection on grounds of cost, reproducibility and injection risk. A skill running under Claude Code has file tools by design, so bounded context is no longer a choice being made. The rejected costs are now real and are accepted: cost per review varies, and a review is not byte-reproducible. The mitigation is `--disallowedTools` on write, network and search tools, so the reviewer can read the repository and talk to the pull request and do nothing else.

**ADR 0002 is superseded.** "One call per pull request" was a question about how to shape API requests. Under an agent loop it is not a decision anyone makes; `--max-turns` bounds the loop instead.

**The offline replay harness in SPEC §5 needs rethinking.** It assumed deterministic inputs. Comparing two versions of a skill now means running both against the same closed pull requests and comparing findings, which is more expensive and less exact. This is the largest thing lost in the change and it is not solved.

What replaced it is weaker and honest: every finding records the reviewer version that produced it (ADR 0006), so two versions can be compared *after the fact* over the pull requests each actually reviewed, rather than *before the fact* over the same ones. That is an observational comparison where the original design promised an experimental one. It needs far more pull requests to say anything, and it cannot answer "what would the old reviewer have said about this diff". SPEC §5 is amended to say so rather than to keep describing a harness nobody is going to build.

**The static-check gate makes the "never report what a linter reports" rule enforceable.** The skill can state it as fact because the tools demonstrably ran on that revision.

**The repository becomes something to copy, not something to install.** Superseded by ADR 0008: the workflows are reusable workflows and the skills are a plugin, so a consuming service holds a stub of about a dozen lines and nothing else.

## What this costs

Precision now depends on prompt text rather than on code, which is harder to test and easier to change carelessly. The metric in `docs/metric.md` is the only thing standing between this design and a skill file that drifts on vibes. It matters more here than it did under the previous design, not less.
