package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Filigree Crawler — Artifact Creature — Insect {4}, 2/2 (Aether Revolt,
// #5536):
//
//	"When this creature dies, create a 1/1 colorless Thopter artifact
//	 creature token with flying."
//
// Thopter Mechanic's death trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f74095aa-232f-432e-a899-86e9463ab5be",
		Name:         "Filigree Crawler",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Filigree Crawler — create a 1/1 colorless Thopter artifact creature token with flying",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 colorless Thopter artifact with flying"), N: 1})),
		},
	})
}
