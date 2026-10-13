package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Exchange of Words — Enchantment {1}{U}{U}:
//
//	"When this enchantment enters, choose two target creatures. For as
//	 long as this enchantment remains on the battlefield, exchange the
//	 text boxes of those creatures."
//
// The one card CR 612.5 names (ADR 0160). Each creature gets the other's
// rules text as it was when the trigger resolved (CR 701.12h) — its
// abilities, its keywords — and keeps its own name, types, power,
// toughness, mana cost and colour. Abilities other effects granted stay
// where they are (CR 612.3).
//
// "Two target creatures" is exactly two, any controller's. If either is
// gone or no longer legal by resolution, no part of the exchange happens
// (CR 701.12a); if the enchantment has already left, the duration is
// over before it begins and nothing happens either. Once made, the
// exchange lasts while the enchantment remains; a creature that leaves
// takes its half with it, and the other keeps its new text until the
// enchantment goes.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a55ef412-257b-4ecd-9440-4545521be938",
		Name:         "Exchange of Words",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Exchange of Words — exchange the text boxes of two target creatures",
					ExchangeTargetTextBoxesWhileSourceRemains),
				TargetCreature("two target creatures").WithCount(2, 2),
			),
		},
	})
}
