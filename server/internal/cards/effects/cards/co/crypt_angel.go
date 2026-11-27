package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Crypt Angel — Creature — Angel {4}{B}, 3/3 (Invasion):
//
//	"Flying, protection from white
//	 When this creature enters, return target blue or red creature card from
//	 your graveyard to your hand."
//
// Moriok Scavenger's enters trigger without the "may", over a creature card
// that is blue or red (either colour, read as printed words, CR 105.2).
// Protection from white is the printed keyword (CR 702.16).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "665f3e52-c536-481b-8b29-fe68f08d1fdf",
		Name:            "Crypt Angel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "protection from white"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Crypt Angel — return target blue or red creature card from your graveyard to your hand", effects.ReturnTargetedCardToHand),
				effects.TargetCardInGraveyard("target blue or red creature card from your graveyard", effects.Or(effects.OfColor("U"), effects.OfColor("R")), effects.Creature(), effects.YouOwn())),
		},
	})
}
