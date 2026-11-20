package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "03a3f54e-9289-4d99-beab-bfa0988a050e",
		Name:            "Guardian Gladewalker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordChangeling},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Guardian Gladewalker — put a +1/+1 counter on target creature",
					PlusOneCounterOnTheTarget),
				TargetCreature("target creature")),
		},
	})
}
