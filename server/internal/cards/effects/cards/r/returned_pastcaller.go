package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Returned Pastcaller — Creature — Spirit Cleric {3}{R}{R/W}{W}, 4/2
// (Strixhaven: School of Mages, #5502):
//
//	"Flying
//	 When this creature enters, return target Spirit, instant, or sorcery card
//	 from your graveyard to your hand."
//
// The target is a card in your graveyard that is a Spirit (a creature or Kindred
// card with the creature type), an instant or a sorcery, chosen as the trigger
// goes on the stack and re-checked as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "58b9f933-7b13-4335-996c-4722ab9e08b2",
		Name:            "Returned Pastcaller",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Returned Pastcaller — return target Spirit, instant, or sorcery card from your graveyard to your hand", effects.ReturnTargetCardToHand),
				effects.TargetCardInGraveyard("target Spirit, instant, or sorcery card from your graveyard", effects.YouOwn(),
					effects.Or(effects.HasSubtype("Spirit"), effects.Instant(), effects.Sorcery()))),
		},
	})
}
