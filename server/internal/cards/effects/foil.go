package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Foil — Instant {2}{U}{U}:
//
//	"You may discard an Island card and another card rather than pay
//	 this spell's mana cost.
//	 Counter target spell."
//
// ADR 0135 §2 (#2412, owner decision 2): two cards under two rules, as a
// set rule on the discard (DiscardEachInstead, TargetSpec.EachOf read
// from the hand). The two picks must fill "an Island card" and "another
// card" one-to-one, so any two cards of which one has the land type
// Island pay it: two Islands do, two non-Islands don't, and an Island
// alone doesn't. Both are discarded as one cost with the spell already
// on the stack.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "7dde4eb6-e9d7-4259-abc2-3af738e0f00f",
		Name:         "Foil",
		Completeness: CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{DiscardEachInstead("an Island card and another card",
			SacrificeSubtype("an Island card", "Island"), AnyCard("another card"))},
		Targets:   TargetSpell("target spell"),
		OnResolve: counterTheTargetSpell,
	})
}
