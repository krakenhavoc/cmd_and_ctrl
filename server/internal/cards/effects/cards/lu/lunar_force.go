package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lunar Force — Enchantment {2}{U} (Eldritch Moon, S82 OLD, #5541):
//
//	"When an opponent casts a spell, sacrifice this enchantment and counter
//	 that spell."
//
// Jin-Gitaxias's cast watch on an opponent's spell, then two instructions in
// printed order: the sacrifice (only if Lunar Force is still on the
// battlefield; it was a different object if it left and came back), then the
// counter of the spell that triggered it. The counter does not target and
// happens even if the sacrifice could not (the spell is simply countered if it
// is still on the stack).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "31b18e10-c74d-4384-a0e2-def0b3ec4799",
		Name:         "Lunar Force",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.AnOpponentCast(nil), "Lunar Force — sacrifice this enchantment and counter that spell",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if !ctx.IsNewSourceObject(ctx.Source()) {
						if err := effects.SacrificeThisIfStillOnBattlefield(g, item); err != nil {
							return err
						}
					}
					return effects.CounterTarget{StackID: ctx.Trigger().Event.CardID}.Apply(ctx)
				}),
		},
	})
}
