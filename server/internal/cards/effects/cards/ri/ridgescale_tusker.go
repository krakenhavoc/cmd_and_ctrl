package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ridgescale Tusker — Creature — Pangolin Beast {3}{G}{G}, 5/5 (Aether Revolt,
// #5536):
//
//	"When this creature enters, put a +1/+1 counter on each other creature
//	 you control."
//
// Web-Warriors' enters trigger: the creatures you control as it
// resolves, the Tusker excluded.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4eef2f24-6a5c-4bc7-b316-dd818dc2741a",
		Name:         "Ridgescale Tusker",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Ridgescale Tusker — put a +1/+1 counter on each other creature you control",
				effects.PutCounterOnEachOtherCreatureYouControl),
		},
	})
}
