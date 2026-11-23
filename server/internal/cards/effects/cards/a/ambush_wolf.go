package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ambush Wolf — Creature — Wolf {2}{G}, 4/2 (Foundations, #4465):
//
//	"Flash
//	 When this creature enters, exile up to one target card from a
//	 graveyard."
//
// Ascendant Dustspeaker's "up to one target card from a graveyard" on an
// enters trigger: with no card chosen it goes on the stack and does
// nothing, and a chosen card that has left its graveyard by resolution
// is not exiled (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4c48de92-e053-4af7-b936-d7ec0875e0d3",
		Name:            "Ambush Wolf",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Ambush Wolf — exile up to one target card from a graveyard", effects.B27ExileChosenTarget),
				effects.TargetCardInGraveyard("up to one target card from a graveyard").WithCount(0, 1)),
		},
	})
}
