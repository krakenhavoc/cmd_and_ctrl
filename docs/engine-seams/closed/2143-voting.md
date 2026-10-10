---
title: "Voting and will of the council"
date: 2026-10-10
issues: [2143]
---
**Voting and will of the council** (#2143, CR 701.38, [ADR 0146](decisions/0146-voting.md)) — `game/council_vote.go` is the rules side, `effects/vote.go` the card side.
- **One option pick per ballot.** `Game.StartVoteForEffect` asks each player in turn order from the named player (CR 701.38a). Each ballot is an ordinary `option_pick` carrying the vote so far (`PendingChoice.CouncilVote`); `ResolveOptionPick` casts it and queues the next, and a voter who leaves casts nothing (the departure table's `dropDefault`). The continuation is a registered key (`RegisterVoteThen`, card side `effects.VoteResultThen`, ledger lines `vote <key>`), so a table waiting on a ballot is a restore point. Readers: `VoteResult.MoreVotes`, `Votes`, `VotersFor`, `MostVotes`.
- **Extra votes (CR 701.38d).** `Spec.ExtraVote`: `game.ExtraVoteYouGet` (Brago's Representative) adds a required ballot, `game.ExtraVoteYouMay` (Ballot Broker, Tivit) an optional one with "Don't vote again". Counted when the player comes to vote, through `CatalogAbilityKey`.
- **Public as it is cast.** `EventVoteCast` is the log's `vote` line; the open ballot carries `council_vote` (options, tally, ballots, `offered`) to every seat.
- **Client.** The ballot's buttons show each option's votes; the other seats see the dock's `voteRequest` read-only (`councilVoteRequest`).
- **Bot.** Cards hint how much the controller and an opponent want each word (`ForController`, `ForOpponents`); an object vote takes the most valuable permanent the bot doesn't control (`aiseat/heuristic/vote.go`).
- **Snapshot.** `pendingChoices[].councilVote`, additive in v7.
- **Cards (9, all Full):** Galadriel, Elven-Queen, Council's Judgment, Expropriate, Plea for Power, Coercive Portal, Magister of Worth, Tivit, Seller of Secrets, Brago's Representative and Ballot Broker.
- **Not built:** secret council (#2926), "whenever players finish voting" (#2927), and the other voting cards, among them Selvala's Stampede (#2928).
