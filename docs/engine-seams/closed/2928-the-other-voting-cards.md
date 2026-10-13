---
title: "The other voting cards"
date: 2026-10-13
issues: [2928]
---
**The other voting cards** (#2928, CR 701.38, [ADR 0146 amendment 2026-10-13](decisions/0146-voting.md), second amendment) — the pieces Selvala's Stampede, Custodi Squire, Trial of a Time Lord, Illusion of Choice and The Valeyard needed on top of the vote.
- **A vote over cards in a graveyard.** `CouncilVote.Zone` / `VotePrompt.OptionZone` / `effects.Vote.Zone` say where the cards the options name are. Empty is the battlefield, as every earlier object vote means; `game.ZoneGraveyard` offers a card only while it is still in a graveyard. Wire: `council_vote.zone`, additive and omitted for every other vote. Snapshot: `pendingChoices[].councilVote.zone`, additive in v8.
- **"You choose how each player votes this turn."** `ModVoteChooser`, a stored scoped-effect rule gate swept at cleanup, read by `voteChooserLocked` at every ballot. The ballot is still the voter's (`CouncilVote.Voter`, the log's `vote` line, the tally) but the prompt is owed to the chooser. The newest record wins; a chooser who leaves stops choosing, and a voter who leaves while the chooser holds their ballot casts nothing.
- **Up to N cards from hand at once.** `PutFromHandOntoBattlefield.Max` and `game.PutCardsFromHandOntoBattlefieldThenForEffect`, the hand twin of the library batch.
- **Reveal until N hits.** `revealUntilCount`, the count form of `revealUntil`.
- **Cards (5):** Selvala's Stampede, Custodi Squire, Trial of a Time Lord, Illusion of Choice (all Full) and The Valeyard (caveat: its villainous-choice ability is not implemented).
- **Not built:** villainous choice (CR 701.55), the keyword action The Valeyard's first ability doubles. Secret council stays #2926.
