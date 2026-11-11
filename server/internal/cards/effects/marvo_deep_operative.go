package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Marvo, Deep Operative — Legendary Creature — Octopus Rogue {3}{U}{B},
// 1/8 (Murders at Karlov Manor Commander, #5424):
//
//	"Whenever Marvo attacks, clash with defending player. (Each clashing
//	 player reveals the top card of their library, then puts that card on
//	 their choice of the top or bottom. A player wins if their card had a
//	 greater mana value.)
//	 Whenever you win a clash, draw a card. Then you may cast a spell from
//	 your hand with mana value 8 or less without paying its mana cost."
//
// The first ability names its opponent, so nothing is chosen: the
// defending player is the player Marvo attacked, or the controller of the
// planeswalker or battle it attacked (defendingPlayerOfAttack), and a
// defending player who has left the game is nobody to clash with. The
// clash is game.ClashForEffect (CR 701.30).
//
// The second watches game.EventClashWon, which names the winner whatever
// started the clash (CR 701.30d), so it also fires on a clash another of
// your cards began, and never on one you lost or tied. The free cast is
// Rishkar's Expertise's (offerFreeHandCast): made as the ability resolves,
// at any timing (CR 608.2g), over the cards in hand with a spell of mana
// value 8 or less, judged on the face being cast. It waits for a prompt
// the draw asks (CR 608.2c), so the card just drawn is one of the choices.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "317ab0b3-1163-4893-a881-4c397feb213d",
		Name:         "Marvo, Deep Operative",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclared(ev, source)
			}, "Marvo, Deep Operative — clash with defending player", func(g *game.Game, item *game.StackItem) error {
				defender := defendingPlayerOfAttack(g, item.Trigger.Event)
				if !defendingPlayerStillIn(g, defender) {
					return nil
				}
				return Clash{Opponent: defender}.Apply(NewContext(g, item))
			}),
			On(game.EventClashWon, youWonAClash,
				"Marvo, Deep Operative — draw a card, then you may cast a spell with mana value 8 or less from your hand without paying its mana cost",
				func(g *game.Game, item *game.StackItem) error {
					return g.DrawNThenForEffect(item.Controller, 1, game.DrawThen{
						Ref: marvoDeepOperativeDrawn, Player: item.Controller, Source: item.SourceCardID,
					})
				}),
		},
	})
}

// marvoDeepOperativeDrawn is "Then you may cast a spell from your hand
// with mana value 8 or less without paying its mana cost." An on-disk
// key: never renamed, never reused.
var marvoDeepOperativeDrawn = game.RegisterDrawThen("marvo-deep-operative/free-cast", func(g *game.Game, d game.DrawThen) error {
	offerFreeHandCastFrom(g, d.Player, d.Source, marvoDeepOperativeFreeCast,
		"a spell with mana value 8 or less", spellWithManaValueAtMost(8))
	return nil
})

// marvoDeepOperativeFreeCast is the offer's keyed continuation (#4246).
var marvoDeepOperativeFreeCast = registerFreeHandCastOffer("choose-cards/marvo-deep-operative-cast", "Marvo, Deep Operative")
