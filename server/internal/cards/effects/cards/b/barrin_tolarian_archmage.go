package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Barrin, Tolarian Archmage — Legendary Creature — Human Wizard {1}{U}{U},
// 2/2 (Jumpstart 2022, #5948):
//
//	"When Barrin enters, return up to one other target creature or
//	 planeswalker to its owner's hand.
//	 At the beginning of your end step, if a permanent was put into your hand
//	 from the battlefield this turn, draw a card."
//
// Miles O'Brien's pair (the draw on your own end step only): the bounce is a
// targeted enters trigger, the draw an intervening "if" (CR 603.4) on the turn's tally of permanents that reached
// your hand from the battlefield (any player's spell or ability, a cost, a
// token bounced before it ceased to exist).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f98f7cb7-90ca-41e7-9f1d-8054931ecf0e",
		Name:         "Barrin, Tolarian Archmage",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Barrin, Tolarian Archmage — return up to one other target creature or planeswalker to its owner's hand",
				effects.BounceEachLegalTarget),
				effects.Another(effects.TargetPermanent("up to one other target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())).WithCount(0, 1))),
			effects.InterveningIf(effects.CondPermanentReturnedToYourHand{},
				effects.AtYourEndStep("Barrin, Tolarian Archmage — draw a card if a permanent was put into your hand from the battlefield this turn",
					effects.Do(effects.DrawCards{N: 1}))),
		},
	})
}
