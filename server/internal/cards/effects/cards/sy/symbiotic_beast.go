package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Symbiotic Beast — Creature — Insect Beast {4}{G}{G}, 4/4 (Dominaria
// Remastered, #5948):
//
//	"When this creature dies, create four 1/1 green Insect creature tokens."
//
// One death trigger; the four tokens are made together (N: 4).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "70ecc25b-2004-4a3c-80c3-6d97bf0711eb",
		Name:         "Symbiotic Beast",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Symbiotic Beast — create four 1/1 green Insect creature tokens",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 green Insect"), N: 4})),
		},
	})
}
