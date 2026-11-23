# Land Luke's six stuck card PRs one at a time

Luke's PRs #6054 (Dissension b1, issue #5598), #6056 (Morningtide b1, #5592), #6057 (Time Spiral b2, #5596), #6058 (Future Sight b1, #5594), #6059 (Planar Chaos b1, #5595) and #6060 (Fifth Dawn b2, #5604) were opened on 2026-11-20, before the #5906 split (PR #5975) landed. Their CI died on the old effects-job out-of-memory and runner-shutdown problem, and #6056 and #6060 now conflict with develop.

Follow s82/QUEUE.md (and through it s82/PLACE.md) on this branch, landing the PRs in the order your prompt gives. These overrides apply:

1. **Trailers.** Use each PR's own sprint and issue: `Sprint: S82 — Card coverage: sets outside Standard, newest first` and `Issue: #<the PR's set issue above>`. Do not use #5948.
2. **Cards develop already has.** Since 2026-11-20 other runs put some of these cards on develop (reprints). After merging develop, find every card file the branch adds whose `OracleID` already appears in develop's tree. For each one, keep develop's card and delete the branch's file, its oracle fixture if develop already has one for that oracle ID, and its rows in the branch's batch test (`*_cards_test.go`). Expected, so check each: 5dn furnace_whelp, skyhunter_prowler; plc sinew_sliver; fut ichor_slick, leaden_fists, street_wraith; tsp assembly_worker, eternity_snare, might_of_old_krosa, sulfurous_blast, tolarian_sentinel, urborg_syphon_mage; mor order_of_the_golden_cricket, spitebellows; dis mistral_charger, rakdos_pit_dragon, taste_for_mayhem. Change the PR's card count in its title or body if it says one.
3. **Token rows.** If a branch adds a token row that develop now has under the same key, keep a single copy. A duplicate map key does not compile.
4. **#6060 lint.** `internal/cards/effects/feedback_bolt.go:28:5: SA4004: the surrounding loop is unconditionally terminated`. Fix it in the moved file without changing behaviour.
5. **Do not break other work.** Change only files the PR itself added or the table files it shares. Never edit another card's behaviour to make a test pass. Never regenerate or delete an existing fixture.
6. **Wall clock.** Stop after 7 hours if the PRs are not all merged.
7. **Last step.** Comment once on each PR you landed, for example "Refreshed onto the split layout; dropped N cards develop already had (...); merged." Do not touch any other PR.
