---
title: "\"Whenever players finish voting\""
date: 2026-10-13
issues: [2927]
---
**"Whenever players finish voting"** (#2927, CR 701.38, [ADR 0146 amendment 2026-10-13](decisions/0146-voting.md)) — `game/council_vote.go` is the rules side, `effects/finish_voting.go` the card side.
- **One event per vote.** `finishVoteLocked` emits `EventVotesFinished` (Actor the vote's controller, Source its card, `Amount` the ballot count) carrying every ballot in `Event.Ballots`, before the vote's own result is applied. A vote that never began (nothing to vote for, `RunVoteThenForEffect`) emits nothing.
- **The ballots ride the event.** A trigger's stack item keeps its event (`ctx.Trigger().Event`), so the ability reads who voted for what as it resolves, and the stack and the event log are both in the snapshot. Additive `ballots` on every `Event` path in the v8 shape file; no new state.
- **The readers.** `game.VotedFor`, `VotedForAChoiceVotedForBy` and `VotedForAChoiceNotVotedForBy`: a player with an extra vote (CR 701.38d) names every choice either vote went to, and one who cast no ballot names none.
- **The card side.** `effects.WheneverPlayersFinishVoting` plus `OpponentsWhoVotedForAChoiceYouVotedFor` / `...YouDidntVoteFor`.
- **Cards.** Grudge Keeper, Model of Unity and Erestor of the Council, all Full.
