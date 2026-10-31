package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cecily, Haunted Mage — Legendary Creature — Human Wizard
// {1}{U}{B}{R}, 3/3:
//
//	"Your maximum hand size is eleven.
//	 Whenever Cecily, Haunted Mage attacks, you draw a card and you lose
//	 1 life. Then if you have eleven or more cards in your hand, you may
//	 cast an instant or sorcery spell from your hand without paying its
//	 mana cost.
//	 Partner—Friends forever"
//
// The maximum is a maximum-hand-size static (ADR 0113 §3, #2074). The
// attack trigger is a Do of three steps (#4228, CR 608.2c), so the life
// loss and the hand count wait for a draw that pauses on a prompt (a dredge
// offer, a draw-instead pick) and for one the life loss asks. The free cast
// is the #3270 pick (offerFreeHandCast, keyed in #4246 so a table waiting on
// it is a restore point): a pass-closed, any-timing {0} cast of
// one instant or sorcery in hand, cast as the trigger resolves (CR
// 608.2g), so a sorcery is castable in the declare attackers step.
// Partner—Friends forever is a deck-construction rule the deck importer
// reads (CR 702.124i).
func init() {
	Register(Spec{
		OracleID:     "a726c27d-2955-4d41-b8a6-fca654c020a2",
		Name:         "Cecily, Haunted Mage",
		Completeness: CompletenessFull,
		HandSize:     []game.HandSizeStatic{YourMaxHandSizeIs(11)},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, ThisAttacked,
				"Cecily, Haunted Mage — draw a card and lose 1 life, then with eleven or more cards in hand cast an instant or sorcery free",
				// The life loss and the free cast wait for a prompt the draw asks (a
				// dredge offer, a draw-instead pick) and for one the life loss asks
				// (CR 608.2c, #4228).
				Do(
					DrawCards{N: 1},
					Instruction(func(ctx *Context) error {
						return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), ctx.Controller(), -1)
					}),
					Instruction(cecilyFreeCast),
				)),
		},
	})
}

// cecilyFreeCast is the last sentence, "Then if you have eleven or more
// cards in your hand, you may cast an instant or sorcery spell from your
// hand without paying its mana cost": the count is read when it runs, after
// the draw and the life loss have finished.
func cecilyFreeCast(ctx *Context) error {
	p := ctx.Game.PlayerByIDForEffect(ctx.Controller())
	if p == nil || p.Hand == nil || len(p.Hand.Cards) < 11 {
		return nil
	}
	offerFreeHandCast(ctx, ctx.Controller(), cecilyFreeCastOffer, "an instant or sorcery spell", freeInstantOrSorcery())
	return nil
}

// cecilyFreeCastOffer is the offer's keyed continuation (#4246), so a
// table waiting on it is a restore point.
var cecilyFreeCastOffer = registerFreeHandCastOffer("choose-cards/cecily-haunted-mage-cast", "Cecily, Haunted Mage")
