package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Martyr of Dusk — Creature — Vampire Soldier {1}{W}, 2/1 (Commander,
// #5368):
//
//	"When this creature dies, create a 1/1 white Vampire creature token
//	 with lifelink."
//
// A dies trigger (graveyard only) written on effects.Do so it declares
// its footprint; the token is the 1/1 white Vampire with lifelink.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "794aaf8a-6d8e-41fb-acbe-a2b73d65d9d9",
		Name:         "Martyr of Dusk",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Martyr of Dusk — create a 1/1 white Vampire creature token with lifelink",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Vampire with lifelink"), N: 1})),
		},
	})
}
