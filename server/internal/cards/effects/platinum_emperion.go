package effects

// Platinum Emperion — Artifact Creature — Golem {8}, 8/8:
//
//	"Your life total can't change. (You can't gain or lose life. You
//	 can't pay any amount of life except 0.)"
//
// The whole card is one clause, and the clause is the DERIVED half of
// #1200 (CR 119.7, CR 119.8, ADR 0085). Declared and no code written:
// "Your life total can't change" is a printed static of a permanent,
// so the engine asks the battlefield on every query
// (game.CatalogPlayerLifeTotalLocked) and writes nothing to the
// player — which is what makes two Emperions compose and one of them
// dying not unlock a total the other is still locking. The same
// argument Leyline of Sanctity's file makes for hexproof, one rule
// over.
//
// The reader asks CatalogAbilityKey, not the printed identity, so an
// Emperion that has lost its abilities (layer 6) stops locking — the
// half of the derivation that a stored flag could not express at all.
//
// WHAT THE PARENTHETICAL IS, precisely, because all three of its
// sentences are load-bearing and all three are implemented:
//
//   - "You can't gain life" — CR 119.7. Every GainLife, every drain's
//     gain half, every lifelink credit, replaced away in the CR 614
//     window #482 routed every life change through.
//   - "You can't lose life" — CR 119.8. A drain, an "each opponent
//     loses 3 life", the life half of an exchange. Damage is still
//     DEALT (so "whenever ~ deals damage" triggers fire and commander
//     damage still accrues toward CR 903.10a) and simply moves no
//     life, which is Platinum Emperion's own printed ruling.
//   - "You can't pay any amount of life except 0" — CR 119.8's second
//     sentence and CR 614.17b. A Phyrexian symbol, a shockland's two
//     life, Snuff Out, an activated ability with a life cost: the COST
//     is refused, at the validator as well as at the payment, so the
//     bot never proposes the line and the client never shows a button
//     that can only fail.
//
// WHAT IT DOES NOT DO, and nobody should expect it to: stop you
// losing the game. CR 104.3 has four routes and this closes only the
// one that runs through your life total — an empty library, ten
// poison counters and 21 commander damage are untouched, and "you
// can't lose the game" is a different clause on different cards
// (tracked on #883). The Emperion is also an 8/8 for {8} that dies to
// everything, which is the rest of why the card is a build-around
// rather than a staple.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:              "bcd2f70c-36c2-44b2-9d4b-1000e9bb62b6",
		Name:                  "Platinum Emperion",
		Completeness:          CompletenessFull,
		PlayerLifeTotalLocked: true,
	})
}
