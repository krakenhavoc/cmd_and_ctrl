package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Judge of Currents — Creature — Merfolk Wizard {1}{W}, 1/1 (Lorwyn):
//
//	"Whenever a Merfolk you control becomes tapped, you may gain 1 life."
//
// Magda's tap watch over Merfolk: two event kinds, because the engine taps an
// attacker without an EventTapCard (a vigilance attacker did not become
// tapped). The Judge is a Merfolk itself, so its own tap pays. The gain is a
// "you may" (CR 603.5).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b4eafd12-5f3e-43e6-aa0a-6bb5f4f5782b",
		Name:         "Judge of Currents",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.OnAny([]game.EventKind{game.EventTapCard, game.EventAttack},
				effects.S82OldB549MerfolkYouControlBecameTapped,
				"Judge of Currents — you may gain 1 life", effects.Do(effects.GainLife{Amount: 1})),
				"Judge of Currents — gain 1 life?"),
		},
	})
}
