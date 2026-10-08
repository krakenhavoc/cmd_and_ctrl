package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Snag — Instant {3}{G}:
//
//	"You may discard a Forest card rather than pay this spell's mana
//	 cost.
//	 Prevent all combat damage that would be dealt by unblocked
//	 creatures this turn."
//
// ADR 0135 §2 (#2412): the alternative cost is a discard
// (DiscardInstead), paid with the spell already on the stack, so a
// discard payoff sees the Forest and a countered Snag does not give it
// back. Any card with the land type Forest pays it, a nonbasic Forest or
// Dryad Arbor included (the ruling).
//
// The shield is #2026's combat status, read as each creature would deal
// combat damage: an unblocked attacking creature's combat damage is
// prevented, and a blocked one's and a blocker's are not.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:         "feb5e9a9-500a-4f20-881d-c97d052953ce",
		Name:             "Snag",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{DiscardInstead("a Forest card", HasSubtype("Forest"))},
		OnResolve:        sourceShieldSpell(combatShieldAgainstCreatures(game.DamageSourceFilter{Combat: game.SourceCombatUnblocked})),
	})
}
