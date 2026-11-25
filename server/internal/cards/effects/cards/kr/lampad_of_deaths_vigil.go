package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lampad of Death's Vigil — Enchantment Creature — Nymph {1}{B}, 1/3 (Theros
// Beyond Death, #5514):
//
//	"{1}, Sacrifice a creature: Each opponent loses 1 life and you gain 1 life."
//
// The sacrifice is a cost (CR 602.2b), any creature you control, the Lampad
// included. Each opponent loses 1 life, then you gain exactly 1 (not the total
// lost): a plain life change from the Lampad, no damage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e0f23e7e-2858-4f5e-9dd5-8a0ba934a048",
		Name:         "Lampad of Death's Vigil",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, Sacrifice a creature: Each opponent loses 1 life and you gain 1 life.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.SacrificeACreature()),
			Effect:  effects.Do(effects.EachOpponentLosesLife{N: 1}, effects.GainLife{Amount: 1}),
		}},
	})
}
