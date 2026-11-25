package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Stoic Builder — Creature — Human {2}{G}, 2/3 (Shadows over Innistrad,
// #5543):
//
//	"When this creature enters, you may return target land card from your
//	 graveyard to your hand."
//
// Circle of the Land Druid's return, made optional: the target is chosen as
// the trigger goes on the stack and re-checked as it resolves (CR 608.2b),
// and the "may" is a real decline.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cca94004-fe8e-404d-9391-ad33d0a126fb",
		Name:         "Stoic Builder",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(
				effects.WhenThisEnters("Stoic Builder — you may return target land card from your graveyard to your hand", effects.S82Mom16ReturnTargetToHand),
				effects.TargetCardInGraveyard("target land card from your graveyard", effects.Land(), effects.YouOwn())),
				"Stoic Builder — return the land card to your hand?"),
		},
	})
}
