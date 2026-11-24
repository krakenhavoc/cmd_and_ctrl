package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Landscaper Colos — Creature — Goat Beast {5}{W}, 4/6 (Modern Horizons 2,
// #5501):
//
//	"When this creature enters, put target card from an opponent's
//	 graveyard on the bottom of their library.
//	 Basic landcycling {1}{W}"
//
// The card goes to the bottom of its owner's library, who is the opponent
// whose graveyard it was in; a card that left in response is an illegal
// target (CR 608.2b). Basic landcycling is CR 702.29e.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "144bef8b-3583-4e98-9285-a1965e7e17f8",
		Name:         "Landscaper Colos",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Landscaper Colos — put target card from an opponent's graveyard on the bottom of their library",
					effects.SocB37PutTargetGraveyardCardOnBottom),
				effects.TargetCardInGraveyard("target card from an opponent's graveyard",
					func(_ *game.Game, caster uuid.UUID, c game.Card) bool { return c.Owner != caster })),
		},
		Activated: []effects.ActivatedAbility{effects.BasicLandcycling("{1}{W}")},
	})
}
