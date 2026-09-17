package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Y'shtola, Night's Blessed — Legendary Creature — Cat Warlock
// {1}{W}{U}{B}, 2/4 (EDHREC rank 2624):
//
//	"Vigilance
//	 At the beginning of each end step, if a player lost 4 or more life
//	 this turn, you draw a card.
//	 Whenever you cast a noncreature spell with mana value 3 or
//	 greater, Y'shtola deals 2 damage to each opponent and you gain 2
//	 life."
//
// The Esper spells commander. Vigilance rides PrintedKeywords. The
// end-step trigger is EVERY end step — any player's turn — with an
// intervening-if over every seated player's life lost this turn
// (b18LifeLostThisTurn: a negative life change or damage to the
// player, the controller's own losses included), checked as the
// trigger would fire and again as it resolves (CR 603.4). The cast
// trigger is Firebrand Archer's condition (b10NoncreatureSpellCastByYou)
// with the mana value read off the stack (game.(*Game).ManaValueForEffect, so an X
// spell counts what was paid); Y'shtola is the damage source, and
// the gain is 2 once, not per opponent.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3268251a-8292-44f9-9267-c961b182f739",
		Name:            "Y'shtola, Night's Blessed",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginEndStep, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b24AnyPlayerLostAtLeastThisTurn(g, 4)
			}, "Y'shtola, Night's Blessed — you draw a card", func(g *game.Game, item *game.StackItem) error {
				if !b24AnyPlayerLostAtLeastThisTurn(g, 4) {
					return nil
				}
				return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
			}),
			On(game.EventCast, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				if !b10NoncreatureSpellCastByYou(ev, source, lki, g) {
					return false
				}
				spell, ok := g.LookupCardForEffect(ev.CardID)
				if !ok {
					return false
				}
				mv, ok := g.ManaValueForEffect(spell)
				return ok && mv >= 3
			}, "Y'shtola, Night's Blessed — 2 damage to each opponent, you gain 2 life", func(g *game.Game, item *game.StackItem) error {
				if err := damageToEachOpponent(g, item, 2); err != nil {
					return err
				}
				return GainLife{Player: item.Controller, Amount: 2}.Apply(NewContext(g, item))
			}),
		},
	})
}
