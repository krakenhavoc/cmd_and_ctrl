package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gravedigger — Creature — Zombie {3}{B}, 2/2 (Foundations Jumpstart):
//
//	"When this creature enters, you may return target creature card from
//	 your graveyard to your hand."
//
// Archaeomancer's return (returnTargetCardToHand) with Amok Time's creature
// target, behind a CR 603.5 "you may". No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1a2030cc-d7ee-4059-b2d7-fb95ea8e267b",
		Name:         "Gravedigger",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(
				effects.WhenThisEnters("Gravedigger — return target creature card from your graveyard to your hand", effects.ReturnTargetCardToHand),
				effects.TargetCardInGraveyard("target creature card from your graveyard", effects.YouOwn(), effects.Creature())),
				"Gravedigger — return it to your hand?"),
		},
	})
}
