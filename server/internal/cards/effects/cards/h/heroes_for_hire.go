package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Heroes for Hire — {3}{R}{R} Enchantment (Marvel Super Heroes Commander,
// #4455):
//
//	"When this enchantment enters, create three Treasure tokens.
//	 Sacrifice a Treasure: Exile the top card of your library. You may play
//	 it this turn."
//
// The second ability is Professional Face-Breaker's (the shared
// sacrificeATreasureToPlayTheTop).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "45e5135b-274c-46b4-8249-7e05e37aaff7",
		Name:         "Heroes for Hire",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Heroes for Hire — create three Treasure tokens",
				effects.Do(effects.CreateToken{Template: effects.TreasureToken(), N: 3})),
		},
		Activated: []effects.ActivatedAbility{
			effects.SacrificeATreasureToPlayTheTop("Sacrifice a Treasure: Exile the top card of your library. You may play it this turn."),
		},
	})
}
