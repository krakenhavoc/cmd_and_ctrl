package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Grenzo, Havoc Raiser — Legendary Creature — Goblin Rogue {R}{R}, 2/2
// (Tarkir: Dragonstorm Commander):
//
//	"Whenever a creature you control deals combat damage to a player,
//	 choose one —
//	 • Goad target creature that player controls.
//	 • Exile the top card of that player's library. Until end of turn,
//	   you may cast that card and you may spend mana as though it were
//	   mana of any color to cast that spell."
//
// One trigger for each creature you control that connects (Toski's
// condition; no "one or more"), the mode chosen as it goes on the stack
// (CR 603.3c). "That player" is the damaged player the trigger's event
// carries. The first bullet's clause is relative to that event
// (ControlledByThatPlayer), so a bullet with no creature to goad is not on
// offer, and a creature that changes hands before resolution is an illegal
// target. The goad is Alela's (the engine ends it at your next turn). The
// second bullet is Breeches's exile-with-permission narrowed to casting
// (a land exiled this way can't be played), for this turn, with mana of
// any color; the card stays in exile when not cast.
//
// No simplification.
func init() {
	t := On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return combatDamageToPlayerBy(ev, source.Controller, g)
	}, "Grenzo, Havoc Raiser — choose one", func(*game.Game, *game.StackItem) error { return nil })
	t.Modes = ChooseOne(
		ModeDoing("Goad target creature that player controls.",
			ControlledByThatPlayer(TargetCreature("target creature that player controls")),
			func(item *game.StackItem, ctx *Context, occ int) error {
				target, ok := ModeTarget(ctx, occ)
				if !ok || target.Kind != game.TargetCard || !onBattlefield(ctx.Game, target.ID) {
					return nil
				}
				b33Goad(ctx.Game, target.ID, item.Controller)
				return nil
			}),
		ModeDoing("Exile the top card of that player's library. Until end of turn, you may cast that card and you may spend mana as though it were mana of any color to cast that spell.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				if item.Trigger == nil {
					return nil
				}
				return ExileTopWithPermission{
					From:     item.Trigger.Event.Target,
					GrantTo:  item.Controller,
					N:        1,
					CastOnly: true,
					AnyColor: true,
				}.Apply(ctx)
			}),
	)
	Register(Spec{
		OracleID:     "80788290-1569-4c7f-b263-b4da7b31a9d1",
		Name:         "Grenzo, Havoc Raiser",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{t},
	})
}
