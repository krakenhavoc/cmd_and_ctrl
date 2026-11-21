package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Barrow Witches — Creature — Human Warlock {4}{B}, 3/4 (Throne of Eldraine, #5516):
//
//	"When this creature enters, return target Knight card from your graveyard to your hand."
//
// Gravedigger's trigger without the "may": the return is mandatory, so the
// trigger is not Optional.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9d4b34cc-0452-42a4-90e4-2a48e7ffadc0",
		Name:         "Barrow Witches",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Barrow Witches — return target Knight card from your graveyard to your hand", effects.ReturnTargetCardToHand),
				effects.TargetCardInGraveyard("target Knight card from your graveyard", effects.YouOwn(), effects.HasSubtype("Knight"))),
		},
	})
}
