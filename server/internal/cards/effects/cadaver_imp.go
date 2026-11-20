package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cadaver Imp — Creature — Imp {1}{B}{B}, 1/1 (Rise of the Eldrazi, #5581):
//
//	"Flying
//	 When this creature enters, you may return target creature card from
//	 your graveyard to your hand."
//
// Gravedigger's optional enters trigger on a flier.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7265d768-b79f-44eb-9873-d834e55f6463",
		Name:            "Cadaver Imp",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Optional(Targeting(
				WhenThisEnters("Cadaver Imp — return target creature card from your graveyard to your hand", returnTargetCardToHand),
				TargetCardInGraveyard("target creature card from your graveyard", YouOwn(), Creature())),
				"Cadaver Imp — return it to your hand?"),
		},
	})
}
