package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Carrier Thrall — Creature — Vampire {1}{B}, 2/1 (Commander Masters):
//
//	"When this creature dies, create a 1/1 colorless Eldrazi Scion creature
//	 token. It has "Sacrifice this token: Add {C}.""
//
// Scion Summoner's printed Eldrazi Scion token (the sacrifice mana ability
// is the token's own) under a plain dies trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b3e0b8a6-5de4-4988-b042-2641b1ed3e32",
		Name:         "Carrier Thrall",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Carrier Thrall — create a 1/1 colorless Eldrazi Scion creature token",
				effects.Do(effects.CreateToken{Template: effects.B28EldraziScionToken(), N: 1})),
		},
	})
}
