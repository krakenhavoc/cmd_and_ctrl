package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wakandan Tusker — Creature — Elephant {1}{G/W}, 2/3 (Marvel Super
// Heroes Commander, #4455):
//
//	"Whenever this creature becomes tapped, you gain 1 life and scry 1."
//
// Dionus's "becomes tapped" (ThisBecameTapped, CR 701.26a): any tap, for
// mana, by an effect or by attacking, and not a vigilance attacker. The
// life is gained first and the scry is asked after it, in printed order.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9e8ffb63-1a7b-4ebe-839d-d5192522db65",
		Name:         "Wakandan Tusker",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, effects.ThisBecameTapped,
				"Wakandan Tusker — you gain 1 life and scry 1", effects.Do(effects.GainLife{Amount: 1}, effects.Scry{N: 1})),
		},
	})
}
