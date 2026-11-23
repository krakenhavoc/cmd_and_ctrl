package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Conscripted Infantry — Creature — Human Soldier {2}{R}, 3/1 (The
// Brothers' War, #5482):
//
//	"When this creature dies, create a 1/1 colorless Soldier artifact
//	 creature token."
//
// The token is an artifact creature, so it feeds every artifact count.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cf0f3aee-78b6-4a98-8e09-8861b73c2e74",
		Name:         "Conscripted Infantry",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Conscripted Infantry — a 1/1 Soldier artifact creature", effects.Do(effects.CreateToken{Template: effects.TokenCard(effects.BroB05SoldierToken), N: 1})),
		},
	})
}
