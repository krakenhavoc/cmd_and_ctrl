package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Startled Relic Sloth — Creature — Sloth Beast {2}{R}{W}, 4/4 (Secrets of
// Strixhaven, #4456):
//
//	"Trample, lifelink
//	 At the beginning of combat on your turn, exile up to one target card
//	 from a graveyard."
//
// Ascendant Dustspeaker's combat trigger: the target is chosen as the
// trigger goes on the stack, "up to one" lets it go on the stack with no
// card, and it is re-checked at resolution (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5d46e85f-4a04-48b1-afe9-3a47678041d4",
		Name:            "Startled Relic Sloth",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample", "lifelink"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				AtBeginningOfYourCombat("Startled Relic Sloth — exile up to one target card from a graveyard",
					b27ExileChosenTarget),
				TargetCardInGraveyard("up to one target card from a graveyard").WithCount(0, 1)),
		},
	})
}
