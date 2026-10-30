# Sprints: milestones, trackers and docs/sprints.md

A sprint lives in three places, and a triage keeps them in step:

1. **A GitHub milestone** titled `S<n> — <theme>`, with an em dash. Its
   description is one or two sentences ending "Opened at the <date> triage."
2. **A tracker issue** with the same title, in that milestone. It is
   labelled `enhancement`, `tier:2-standard` and `P2`, and assigned to no one.
   Its body is the checklist.
3. **`docs/sprints.md`**: a row in the sprint index and a section before
   `## Backlog`. Every change to it goes in one `docs:` PR per triage.

`scripts/sprints_md.py` writes all of this from one spec, so the three match.
It reproduces the S76 to S80 changes of 2026-10-30 byte for byte.

## Opening a sprint

1. **Number.** Take the next unused one. Check the sprint index in
   `git show origin/develop:docs/sprints.md` and every milestone, open and
   closed (`query.py milestones` lists the open ones; `milestones.json` in
   the snapshot has all of them).
2. **Spec.** Write `$TRIAGE_DIR/s81.json`:

   ```json
   {
     "key": "S81",
     "name": "Theme: the few words a milestone list can show",
     "phase": 7,
     "tracker": null,
     "date": "2026-11-13",
     "goal": "what done looks like, in one or two sentences",
     "context": "optional: where these issues came from, what is deliberately left out",
     "groups": [
       {"items": [
         {"n": 4326, "text": "select several attackers, then one target for all of them (player report)"},
         {"n": 4544, "text": "extort plays the Scales of Souls (PR #4551, merged into `develop`)", "done": true}
       ]},
       {"heading": "Seams the batches found:", "items": [
         {"n": 4586, "text": "sneak (PR #4644 open)"},
         {"text": "The later sets, #4458 to #4471, wait in this milestone in #4452's order."}
       ]}
     ],
     "exit_criteria": [
       "Something a person can check, such as a test, a game on cmd-dev or a screen size.",
       "Every listed seam is closed, or parked with its reason."
     ]
   }
   ```

   - Item text is a short lower-case summary, not the issue title.
   - Mark `"done": true` for issues already closed or merged.
   - Exit criteria are checkable outcomes, not tasks.
   - Optional fields: `opened` replaces "Opened at the <date> triage.";
     `status` replaces "**Planned** (<date> triage)."; `row_status` replaces
     `planned` in the index (for example `in progress`).
3. **Tracker body:**
   `sprints_md.py render $TRIAGE_DIR/s81.json --part tracker > $TRIAGE_DIR/bodies/s81.md`
4. **Milestone and tracker:**

   ```sh
   new-sprint.sh "S81 — Theme: the few words" $TRIAGE_DIR/bodies/s81.md \
     "What it covers, in a sentence. Opened at the 2026-11-13 triage."
   ```

   It reuses an existing milestone or tracker with that title, so a rerun is
   safe. Put the tracker number it prints into the spec's `"tracker"`.
5. **Issues.** Set the sprint's issues into the milestone with `edit.sh`.
6. **docs/sprints.md**, in the docs worktree (below):
   `sprints_md.py add-sprint docs/sprints.md $TRIAGE_DIR/s81.json --dry-run`,
   read the diff, then run it without `--dry-run`.

## Adding issues to an existing sprint

- Append an "Added at the <date> triage:" block to the tracker body
  ([templates.md](templates.md)).
- In `docs/sprints.md`, add a sentence to that sprint's Status naming the
  issues and why they fit the goal. For example: "The 2026-10-30 triage added
  three heuristic-bot bugs under the goal's "sharper bots": [#4320](…) (…),
  …". Use `sprints_md.py apply` with an `old`/`new` or `after`/`insert` edit
  anchored on the paragraph's last sentence.
- In the doc, link issues as `[#n](https://github.com/<repo>/issues/n)`. In
  issue bodies and comments, plain `#n` is enough.

## Finishing a sprint

When every checklist issue is closed (check each with `gh issue view <n>
--json state`, not the ticks):

1. Close the tracker with the comment in [templates.md](templates.md).
2. `sprints_md.py mark-done docs/sprints.md S76 2026-10-30 "Every checklist
   issue is closed, the last two being …"`. The index row gets the date and
   **done**. The Status gets "**Done** (<date> triage). <note>", and the old
   status stays below as "Before that: …".

## Parking in the Backlog

The Backlog section ends with prose naming each parked issue and what it waits
on. Add a sentence before its last one ("Pull one into a sprint when its
dependency clears."): "The <date> triage parked [#n](…) (<what>), which waits
on <what>, and …".

## The docs PR

Keep every `docs/sprints.md` change of the run in one PR.

```sh
git fetch origin
git worktree add .claude/worktrees/plan-s81 -b docs/plan-s81 origin/develop
cd .claude/worktrees/plan-s81
# sprints_md.py add-sprint / mark-done / apply against docs/sprints.md
git add docs/sprints.md
git commit -m "docs: plan S81 at the 2026-11-13 triage"   # plus a body saying why
git push -u origin docs/plan-s81
gh pr create --base develop --title "docs: plan S81 at the 2026-11-13 triage" --body-file <body>
```

- The title and commit are `docs: <subject>`, with no scope: never
  `docs(docs):`. Use Conventional Commits, and add no Claude attribution.
- The body follows the template in [templates.md](templates.md). Use `Refs
  #<tracker>`: a `Closes` line would close the tracker on promotion.
- Always pass `--base develop`. `main` is the default branch, and
  `promotion-guard` fails a feature PR into it.
- Wait for CI with `ci-status.sh <pr>` about every 90 seconds. Exit 2 means
  still running and 3 means rate-limited (not red). A docs-only PR still runs
  the server-test shards, about 10 minutes.
- A failed `server-check` whose log says `Timeout exceeded` is the
  golangci-lint flake: `gh run rerun <run-id> --failed`.
- When CI is green, `gh pr merge <pr> --squash`. Merging into `develop` has
  standing approval; never merge into `main`, and never use `--admin`.
- Then clean up from the main checkout:
  `git worktree remove .claude/worktrees/plan-s81` and
  `git branch -D docs/plan-s81`.
