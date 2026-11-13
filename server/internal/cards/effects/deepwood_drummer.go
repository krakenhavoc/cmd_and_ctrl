package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Deepwood Drummer — Creature — Human Spellshaper {1}{G}, 1/1
// (Mercadian Masques, #5619):
//
//	"{G}, {T}, Discard a card: Target creature gets +2/+2 until end of
//	 turn."
//
// The spellshaper cost (Kris Mage's: mana, {T} and DiscardACard, the
// card discarded as the ability is activated, CR 602.2b) on Tuknir
// Deathlock's shared boostTargetUntilEOT body at +2/+2, declared a
// benefit to the target. The boost ends in the cleanup step (CR 514.2).
// The {T} is a creature's, so the Drummer waits out summoning sickness
// (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b2bbb4e5-05b1-49e5-ad02-7384e115e06b",
		Name:         "Deepwood Drummer",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: "{G}, {T}, Discard a card: Target creature gets +2/+2 until end of turn.",
			Purpose: game.Purpose{
				Answers: game.AnswerPump,
				Targets: game.ForTargets(BenefitUntilEOT(0, 2, 2)),
			},
			Cost:    Plus(ManaCost("{G}"), TapCost(), DiscardACard()),
			Targets: TargetCreature("target creature"),
			Effect:  boostTargetUntilEOT("Deepwood Drummer — +2/+2", 2, 2),
		}},
	})
}
