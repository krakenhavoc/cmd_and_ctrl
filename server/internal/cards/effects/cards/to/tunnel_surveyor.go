package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tunnel Surveyor — Creature — Human Detective {2}{U}, 2/2 (Duskmourn:
// House of Horror, #4466):
//
//	"When this creature enters, create a 1/1 white Glimmer enchantment
//	 creature token."
//
// An enters trigger making Grand Entryway's Glimmer token.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7aae54a6-3011-4867-b5ee-21a2111aecc6",
		Name:         "Tunnel Surveyor",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Tunnel Surveyor — create a 1/1 white Glimmer enchantment creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Glimmer enchantment"), N: 1})),
		},
	})
}
