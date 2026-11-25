package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Defiant Greatmaw — Creature — Hippo {2}{G}, 4/5 (Amonkhet, S82 OLD batch 242, #5535):
//
//	"When this creature enters, put two -1/-1 counters on target creature you
//	 control.
//	 Whenever you put one or more -1/-1 counters on this creature, remove a
//	 -1/-1 counter from another target creature you control."
//
// The second trigger is once per batch of -1/-1 counters you put on the
// Greatmaw (CR 603.2c), whether from its own enters trigger or another source;
// counters an opponent put don't count. Its target is another creature you
// control, chosen as it goes on the stack, and a creature with no -1/-1 counter
// loses nothing.
//
// No simplification.
func init() {
	remove := effects.WheneverYouPutCountersOfKindOnACreature(game.CounterMinusOne,
		func(_ *game.Game, source *game.Card, c game.Card) bool { return c.InstanceID == source.InstanceID },
		"Defiant Greatmaw — remove a -1/-1 counter from another target creature you control",
		func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			target, ok := ctx.ClauseTarget(0)
			if !ok || target.Kind != game.TargetCard {
				return nil
			}
			if c, ok := g.LookupCardForEffect(target.ID); !ok || c.Counters[game.CounterMinusOne] <= 0 {
				return nil
			}
			return effects.AddCounter{Target: target.ID, Kind: game.CounterMinusOne, N: -1}.Apply(ctx)
		})
	remove.Targets = effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl()))
	effects.Register(effects.Spec{
		OracleID:     "ee152ffa-9c39-48ed-9aca-e53c8ebb0231",
		Name:         "Defiant Greatmaw",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OldB242WhenThisEntersMinusCountersOnYourCreature("Defiant Greatmaw", 2),
			remove,
		},
	})
}
