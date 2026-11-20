package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Corpse Cur — Artifact Creature — Phyrexian Dog {4}, 2/2 (Scars of
// Mirrodin, #5577):
//
//	"Infect
//	 When this creature enters, you may return target creature card with
//	 infect from your graveyard to your hand."
//
// Gravedigger's optional enters trigger with the clause narrowed to a
// creature card that has infect. Infect is the engine's damage tail
// (CR 702.90).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8f59e4d9-67ba-4dc6-ae3e-183e63673318",
		Name:            "Corpse Cur",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"infect"},
		Triggered: []game.TriggeredAbility{
			Optional(Targeting(
				WhenThisEnters("Corpse Cur — return target creature card with infect from your graveyard to your hand", returnTargetCardToHand),
				TargetCardInGraveyard("target creature card with infect from your graveyard", YouOwn(), Creature(), HasKeyword("infect"))),
				"Corpse Cur — return it to your hand?"),
		},
	})
}
