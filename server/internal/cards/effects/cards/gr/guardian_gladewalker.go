package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Guardian Gladewalker — Creature — Shapeshifter {1}{G}, 1/1 (Kaldheim,
// #5505):
//
//	"Changeling (This card is every creature type.)
//	 When this creature enters, put a +1/+1 counter on target creature."
//
// Changeling rides PrintedKeywords (Barkform Harvester); the enters
// trigger is Backup Agent's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "03a3f54e-9289-4d99-beab-bfa0988a050e",
		Name:            "Guardian Gladewalker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordChangeling},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Guardian Gladewalker — put a +1/+1 counter on target creature",
					effects.PlusOneCounterOnTheTarget),
				effects.TargetCreature("target creature")),
		},
	})
}
