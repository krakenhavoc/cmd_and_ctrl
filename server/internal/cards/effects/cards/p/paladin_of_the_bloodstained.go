package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Paladin of the Bloodstained — Creature — Vampire Knight {3}{W}, 3/2
// (Foundations Jumpstart):
//
//	"When this creature enters, create a 1/1 white Vampire creature token
//	 with lifelink."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f340018d-8b68-45d5-a72e-a922be68365d",
		Name:         "Paladin of the Bloodstained",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Paladin of the Bloodstained — create a 1/1 white Vampire creature token with lifelink",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Vampire with lifelink"), N: 1})),
		},
	})
}
