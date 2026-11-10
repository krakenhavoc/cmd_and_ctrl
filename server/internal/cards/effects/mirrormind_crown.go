package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mirrormind Crown — Artifact — Equipment {4} (Lorwyn Eclipsed, #4458):
//
//	"As long as this Equipment is attached to a creature, the first time
//	 you would create one or more tokens each turn, you may instead create
//	 that many tokens that are copies of equipped creature.
//	 Equip {2}"
//
// Moonlit Meditation's replacement (copiesOfAttachedInsteadOfFirstTokens)
// on the equipped creature: an optional "instead" (CR 614.1a) on the
// creation instruction (CR 701.7b) that keeps each group's count and
// swaps the kind for a copy of the equipped creature, taken as the
// instruction is replaced. The copies take only its copiable values
// (CR 707.2) — not its counters, tapped state, attachments or non-copy
// effects — and a creature that is itself a copy or a token is copied as
// the rulings say (TokenCopyTemplate). The copies enter as new objects,
// so their enters abilities trigger and their "as this enters" choices
// are made.
//
// "As long as this Equipment is attached to a creature" is the host
// check: the Crown does nothing while unattached, or attached to a
// permanent that is no longer a creature. "The first time each turn" is
// the first creation instruction under your control this turn whether
// or not the Crown was there for it; declining uses it up.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "353db389-b829-4e66-85d5-6018bc87c3e0",
		Name:         "Mirrormind Crown",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{copiesOfAttachedInsteadOfFirstTokens(
			"Mirrormind Crown: create copies of the equipped creature instead",
			"Mirrormind Crown — create copies of the equipped creature instead?",
			func(host game.Card) bool { return host.IsCreature() },
		)},
		Activated: []ActivatedAbility{EquipAbility("{2}")},
	})
}
