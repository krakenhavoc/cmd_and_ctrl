package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scrib Nibblers — Creature — Rat {2}{B} 1/1 (Worldwake, S82 OLD batch 489,
// #5582):
//
//	"{T}: Exile the top card of target player's library. If it's a land
//	 card, you gain 1 life.
//	 Landfall — Whenever a land you control enters, you may untap this
//	 creature."
//
// The exile is one card off the top of the chosen player's library, and the
// life is gained only if that card is a land card. Landfall is optional.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "458800d3-47b7-4f28-8371-25c8d1e60011",
		Name:         "Scrib Nibblers",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{LifeGain: 1},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Exile the top card of target player's library. If it's a land card, you gain 1 life.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetPlayer("target player"),
			Effect:  effects.S82OldB489ExileTopCardOfTargetPlayerGainIfLand,
		}},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Landfall("Scrib Nibblers — you may untap this creature", effects.B33UntapSelf),
				"Scrib Nibblers — untap it?"),
		},
	})
}
