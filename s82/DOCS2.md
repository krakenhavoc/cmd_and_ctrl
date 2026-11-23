# Follow-ups from the docs sync (#6079), approved by Ian on 2026-11-23

Read AGENTS.md first. Work on a fresh branch off `origin/develop` named `docs/sync-followups-2026-11`. Ian approved all four items below. Do only these.

1. **ADRs to Accepted.**
   - Set the status of ADRs 0224, 0114, 0115, 0036, 0122, 0161, 0162, 0093, 0077 and 0052 from Proposed to Accepted.
   - Check 0113 against the code. If the code implements it, set it to Accepted too. If not, leave it and say why.
   - Update each row in docs/decisions/README.md to match.
   - Change only the status line. Never change a decision's text.
2. **Uniform status wording.** Change every ADR status written as "Implemented" or lowercase "accepted" to "Accepted", in the files and in the README index. Leave "Superseded", "Rejected" and the like alone.
3. **Small fixes.**
   - In the 17 fragments under `docs/engine-seams/closed/` that link `../../decisions/…`, change the link to the form that works once the fragment is rendered into `docs/engine-seams.md` (that is, `decisions/…`). Regenerate `docs/engine-seams.md` with the roadmap generator, never by hand. If the generator or a docsguard link test wants a different form, follow the test.
   - In `server/Makefile`, the `.PHONY` line reads `buildbuild-bot`. Split it into the two intended targets.
4. **Close finished sprints.**
   - Read docs/sprints.md. For S79 (tracker #4649) and S80, confirm every checklist item on the tracker issue is done or merged.
   - Close the tracker with a one-line comment: "All slices merged; closed with Ian's OK, 2026-11-23."
   - Mark the sprint closed in docs/sprints.md.
   - Leave epics #4399, #4870 and #5431 open. Leave S70 and S81 open.
   - If a tracker has an unticked item that is not done, do not close it. List it.

Then:
- Run `go test ./internal/docsguard/ ./internal/roadmap/`.
- Commit and open the PR per AGENTS.md §4.
  - Sprint: S82 — Card coverage: sets outside Standard, newest first.
  - Open an issue "Docs sync follow-ups, November 2026" (`tier:1-mechanical`) and close it from the PR.
- `--base develop`, then `gh pr merge --squash --auto`, and wait for CI.
  - If develop moves and conflicts, merge develop in. Plain push, never force.
  - Rerun failures that are only infrastructure (OOM, exit 137/143, runner shutdown).
- Stop when it merges, or after 3 hours.

Final message: the PR number, whether it merged, which ADRs changed, which trackers closed, and anything you left and why.

Nobody will answer you mid-run. Do not end your turn while CI is running.
