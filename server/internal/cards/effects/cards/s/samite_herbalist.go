package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Samite Herbalist — Creature — Human Cleric {1}{W}, 2/1 (Dominaria
// United):
//
//	"Whenever this creature becomes tapped, you gain 1 life and scry 1."
//
// Compassionate Healer's trigger: "becomes tapped" is the tap event and the
// attack that taps it (ThisBecameTapped, #4534), so a vigilance attack is
// not one. The scry is declared on the card's Purpose.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "af753429-c121-4957-b768-8bed43dc06cb",
		Name:         "Samite Herbalist",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Scry: 1},
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, effects.ThisBecameTapped,
				"Samite Herbalist — you gain 1 life and scry 1",
				effects.Do(effects.GainLife{Amount: 1}, effects.Scry{N: 1})),
		},
	})
}
