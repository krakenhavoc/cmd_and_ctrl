package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Robber of the Rich — Creature — Human Archer Rogue {1}{R}, 2/2:
//
//	"Reach, haste
//	 Whenever this creature attacks, if defending player has more cards
//	 in hand than you, exile the top card of their library. During any
//	 turn you attacked with a Rogue, you may cast that card and you may
//	 spend mana as though it were mana of any color to cast that spell."
//
// An intervening "if" (CR 603.4): the defending player must have more
// cards in hand than the Robber's controller both as the attack is
// declared and as the trigger resolves. "Defending player" is the player
// attacked, or the controller of the planeswalker or battle attacked.
//
// The permission is a conditional exile play (#3192, ADR 0066's second
// 2026-10-13 amendment) held by the Robber's controller over the
// opponent's card: it lasts for as long as the card stays exiled and is
// live only during a turn its holder attacked with a Rogue
// (game.CondAttackedWithSubtype, read off the attacker as it was
// declared; a changeling is every creature type). It is a CAST only, so
// a land exiled this way is stranded, and any-colour mana pays for it.
//
// No simplification.
func init() {
	moreCardsInHand := func(g *game.Game, controller, defender uuid.UUID) bool {
		d, me := g.PlayerByIDForEffect(defender), g.PlayerByIDForEffect(controller)
		return d != nil && me != nil && d.Hand.Size() > me.Hand.Size()
	}
	Register(Spec{
		OracleID:        "4df0d381-fc05-4d30-9307-48963aa2ecfd",
		Name:            "Robber of the Rich",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "haste"},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attackDeclared(ev, source) && moreCardsInHand(g, source.Controller, defendingPlayerOfAttack(g, ev))
			}, "Robber of the Rich — exile the top card of defending player's library", func(g *game.Game, item *game.StackItem) error {
				defender := defendingPlayerOfAttack(g, item.Trigger.Event)
				if !moreCardsInHand(g, item.Controller, defender) {
					return nil
				}
				return ExileTopWithPermission{
					From: defender, GrantTo: item.Controller, N: 1,
					CastOnly: true, AnyColor: true, WhileExiled: true,
					Condition: &game.PermissionCondition{Key: game.CondAttackedWithSubtype, Subtype: "Rogue"},
				}.Apply(NewContext(g, item))
			}),
		},
	})
}
