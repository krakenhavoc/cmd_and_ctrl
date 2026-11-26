package cardslu

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"

// Master Decoy — Creature — Human Soldier {1}{W}, 1/2 (Duel Decks: Speed vs.
// Cunning, #5555):
//
//	"{W}, {T}: Tap target creature."
//
// Akroan Mastiff's tapper.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "94a16609-8529-4301-8a3d-47ced194f7cd",
		Name:         "Master Decoy",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}, {T}: Tap target creature.",
			Cost:    effects.Plus(effects.ManaCost("{W}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.TapChosenPermanent,
		}},
	})
}
