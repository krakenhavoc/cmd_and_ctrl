package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Braulios of Pheres Band — Legendary Creature — Centaur Scout {3}{G}{G},
// */* (Foundations Jumpstart):
//
//	"Braulios's power and toughness are each equal to the number of lands
//	 you control.
//	 Whenever Braulios attacks, draw a card, then you may put a land card
//	 from your hand onto the battlefield."
//
// Beanstalk Giant's characteristic-defining P/T (layer 7a) and Chulane's
// draw-then-put. The land is put, not played (CR 305.4), so it is not a land
// drop.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "70e247e5-6ca4-458c-9519-d19b400955f6",
		Name:         "Braulios of Pheres Band",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Lands: 1, LandsUntapped: 1},
		Static: []game.StaticAbility{
			effects.LandsYouControlPT("Braulios's power and toughness are each equal to the number of lands you control."),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, effects.Self, "Braulios of Pheres Band — draw a card, then you may put a land card from your hand onto the battlefield",
				effects.Do(effects.DrawCards{N: 1}, effects.MayPutALandFromHand("Braulios of Pheres Band"))),
		},
	})
}
