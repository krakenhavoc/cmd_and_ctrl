package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Voldaren Epicure — Creature — Vampire {R}, 1/1 (Innistrad: Crimson Vow):
//
//	"When this creature enters, it deals 1 damage to each opponent. Create a
//	 Blood token."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cef82674-1f61-407b-801b-a76a3803e598",
		Name:         "Voldaren Epicure",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Voldaren Epicure — 1 damage to each opponent and create a Blood token",
				effects.Do(effects.DamageEachOpponent{N: 1}, effects.S82Vow03CreateBlood(1))),
		},
	})
}
