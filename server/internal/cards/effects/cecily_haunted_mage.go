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
// attack trigger draws through Game.DrawNThenForEffect, so the life loss
// and the hand count wait for a draw that pauses on a prompt (a dredge
// offer, a draw-instead pick). The free cast is the #3270 pick
// (offerFreeHandCastFrom, #2110): a pass-closed, any-timing {0} cast of
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
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return g.DrawNThenForEffect(ctx.Controller(), 1, game.DrawThen{
						Ref:    cecilyAfterDraw,
						Player: ctx.Controller(),
						Source: ctx.Source(),
					})
				}),
		},
	})
}

// cecilyAfterDraw is the rest of the trigger, once the draw is done:
// the life loss, then the intervening count and the free-cast pick.
var cecilyAfterDraw = game.RegisterDrawThen("cecily-haunted-mage-after-draw", func(g *game.Game, d game.DrawThen) error {
	if err := g.ChangePlayerLifeForEffect(d.Source, d.Player, -1); err != nil {
		return err
	}
	p := g.PlayerByIDForEffect(d.Player)
	if p == nil || p.Hand == nil || len(p.Hand.Cards) < 11 {
		return nil
	}
	offerFreeHandCastFrom(g, d.Player, d.Source, "Cecily, Haunted Mage", "an instant or sorcery spell", Or(Instant(), Sorcery()))
	return nil
})
