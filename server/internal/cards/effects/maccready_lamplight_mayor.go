package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// MacCready, Lamplight Mayor — Legendary Creature — Human Advisor {W}{B},
// 1/3 (Fallout, #5948):
//
//	"Whenever a creature you control with power 2 or less attacks, it gains
//	 skulk until end of turn. (It can't be blocked by creatures with
//	 greater power.)
//	 Whenever a creature with power 4 or greater attacks you, its
//	 controller loses 2 life and you gain 2 life."
//
// Two printed triggers, two rows, one per attacking creature. Power is read
// as the attack is declared (layers applied), the first trigger's creature
// being any creature MacCready's controller controls (MacCready himself
// included) and given skulk until end of turn if it is still on the
// battlefield as the ability resolves. The second is any player's creature
// attacking MacCready's controller themselves (CR 508.3e; a planeswalker
// or battle they protect does not count): the attacking creature's
// controller, read as it attacked, loses 2 life, then MacCready's controller
// gains 2 life, whether or not the creature is still there.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f40a1a41-d340-4256-8f97-77499cd511a8",
		Name:         "MacCready, Lamplight Mayor",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attackDeclaredByYou(ev, source.Controller) && b31CurrentPowerOf(g, ev.CardID) <= 2
			}, "MacCready, Lamplight Mayor — the creature gains skulk until end of turn", func(g *game.Game, item *game.StackItem) error {
				id := item.Trigger.Event.CardID
				if !onBattlefield(g, id) {
					return nil
				}
				return untilEndOfTurn(NewContext(g, item), id, nil, "MacCready, Lamplight Mayor — skulk until end of turn",
					game.AddKeywordsMod("skulk"))
			}),
			On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Kind == game.EventAttack && ev.Target == source.Controller && ev.Actor != uuid.Nil &&
					ev.Actor != source.Controller && b31CurrentPowerOf(g, ev.CardID) >= 4
			}, "MacCready, Lamplight Mayor — its controller loses 2 life and you gain 2 life", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if err := g.ChangePlayerLifeForEffect(item.SourceCardID, item.Trigger.Event.Actor, -2); err != nil {
					return err
				}
				return GainLife{Player: ctx.Controller(), Amount: 2}.Apply(ctx)
			}),
		},
	})
}
