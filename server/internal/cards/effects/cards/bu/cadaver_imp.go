package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "7265d768-b79f-44eb-9873-d834e55f6463",
		Name:            "Cadaver Imp",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(
				effects.WhenThisEnters("Cadaver Imp — return target creature card from your graveyard to your hand", effects.ReturnTargetCardToHand),
				effects.TargetCardInGraveyard("target creature card from your graveyard", effects.YouOwn(), effects.Creature())),
				"Cadaver Imp — return it to your hand?"),
		},
	})
}
