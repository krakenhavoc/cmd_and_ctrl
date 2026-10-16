# ADR 0199 — A hand played revealed

**Status:** Accepted · 2026-10-16 · S76
**Issue:** [#3953](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3953) (split from [#3531](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3531); Revelation is from [#1997](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1997))
**Numbering:** every remote head was listed (`develop`, `main`, `cost-ledger`, `docs/issue-audit`, `feat/3868-tap-additional-cost`, `feat/3872-text-changing-readers`, `feat/3928-activate-as-though-haste`, `feat/750-conditional-block-restrictions`, `feat/playmats`, `fix/2947-resolution-copy-casts`, `fix/2950-darkness-replacement-continuation`, `fix/3870-movecost-mill`, `fix/caddy-reload-admin-off`, `screenshots/3934-playmat`, `wip/836-one-click-default`, `pr/2326`). The highest in use is **0198**; 0190 is a gap ADR 0197 already left alone, so this is **0199**.
**Builds on:** [ADR 0066](0066-granted-cast-and-play-permissions.md) decision 5 (the top of a library, visible by a rule derived per frame), [ADR 0028](0028-admin-context-menu.md)'s 2026-10-13 amendment (`KnownBy` and `known_by_seats` on a hand card), [ADR 0063](0063-durations-and-control.md) (the CR 611.2 duration model) and the player statics of [ADR 0072](0072-protection.md)'s 2026-09-22 amendment.

---

## Context

Three cards make a player play with their hand revealed:

- Telepathy: "Your opponents play with their hands revealed."
- Revelation (a world enchantment): "Players play with their hands revealed."
- Stromgald Spy: "Whenever this creature attacks and isn't blocked, you may have defending player play with their hand revealed for as long as this creature remains on the battlefield. If you do, this creature assigns no combat damage this turn."

The engine had one kind of hand reveal: a one-time CR 701.20a reveal, which stamps each revealed card's `KnownBy` with every seat. That is the wrong shape for these cards. Their effect covers whatever is in the hand at any moment while it lasts, including a card drawn later (CR 611.3a for the two statics; the Spy's "for as long as" is CR 611.2b), and it ends: once Telepathy leaves or the Spy leaves, the hand is hidden again. CR 400.2 keeps the hand a hidden zone throughout ("even if all the cards in one such zone happen to be revealed"), so nothing about the zone itself changes, only who may see its cards. CR 701.20c's example uses Telepathy: a card already revealed this way may still be revealed again to pay a cost.

## Decision

### 1. A revealed hand is a per-player state, derived where it is read

`Game.HandRevealedLocked(player)` (`game/hand_revealed.go`) answers "does this player play with their hand revealed right now, and why", as a list of source names. Nothing is written into `Card.KnownBy`. A stamp per card would have to be added as each card enters the hand, removed as it leaves, and removed from every card when the effect ends, which is the argument ADR 0066 decision 5 made for the top of a library. Deriving it means a card drawn while the effect lasts is shown, and the hand is hidden again the moment the effect ends, with nothing to invalidate.

Revealed means revealed to every player (CR 701.20a), so the state has no per-viewer half. Telepathy reveals each opponent's hand to all players, including the other opponents, not just to its controller.

### 2. Two sources, as every player-level state has

- **Derived**: a permanent's static, declared as `Spec.HandsRevealed` (`CardDef.HandsRevealed`, read through `CatalogHandsRevealed`): `HandRevealOpponents` for Telepathy, `HandRevealEachPlayer` for Revelation. It is read off the permanent's ability key, so a Telepathy that has lost its abilities reveals nothing. Every other player is an opponent; the engine has no teams.
- **Granted**: `PlayerStatic.HandRevealed`, stored on the player whose hand is revealed, with a CR 611.2 `Duration` (`Game.GrantHandRevealedForEffect`). Stromgald Spy grants it with `DurationWhileSourceRemains`, so a Spy that left before its trigger resolved never starts the duration (CR 611.2b) and reveals nothing. The reader tests the duration live, as `playerAbilityTokensLocked` does, so the hand is hidden again the instant the Spy leaves, between sweeps.

### 3. The view makes every card in the hand known to every seat

`stampHandRevealed` (`protocol/view.go`) runs beside `stampLibraryTop`, before the per-viewer redaction: every card in a revealed hand gets every seat in its `knowers` set. The existing opponent-hand projection (`keepKnownInHandZone`) then keeps those cards face up for each other seat, a spectator sees them because every seat knows them, and the public log names them while the effect lasts. The owner's own cards carry `known_by_seats` as they would after any reveal. Each viewer's cast stamps on those cards are stripped as before (#1166), so nothing about the owner's own options leaks.

`PlayerView.hand_revealed` names the sources ("Telepathy", "Stromgald Spy"). It is public, and the client shows it as a HAND OPEN badge on the seat with the sources in its tooltip, so the table knows why that hand is face up. The owner's per-card known-by counts are hidden while it is set, since the badge already says every card is known.

A borrowed ability's lender in a revealed hand is named to every seat (`LenderKnowersLocked`, [ADR 0155](0155-having-another-objects-activated-abilities.md) decision 10).

### 4. Snapshot

`PlayerStatic.HandRevealed` is one new `omitempty` bool on an existing snapshot struct, an additive change within the current version (#522). A binary from before it reads the grant back as an entry that says nothing, so after a rollback the hand is hidden early: a missing reveal, never an extra one. The derived half is not stored.

## Consequences

- Telepathy, Revelation and Stromgald Spy ship with no simplifications.
- Any later "play with your hand revealed" card is one `Spec.HandsRevealed` value or one `GrantHandRevealedForEffect` call.
- The bot sees a revealed hand the way it sees any revealed card: through its seat's frame.

## Out of scope

- "Look at target player's hand" is a one-time private look and keeps its existing path.
- Field of Dreams' "Players play with the top card of their libraries revealed" is about libraries; it stays on #1997.
