package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bloodcrazed Hoplite — Creature — Human Soldier {1}{B}, 2/1:
//
//	"Heroic — Whenever you cast a spell that targets this creature,
//	 put a +1/+1 counter on it.
//	 Whenever a +1/+1 counter is put on this creature, remove a +1/+1
//	 counter from target creature an opponent controls."
//
// #1841: the removal triggers once per counter put on the Hoplite
// (CR 603.2c), each with its own target, so a Hardened Scales'd heroic
// takes two counters. The target is chosen as the trigger goes on the
// stack; with no creature an opponent controls the trigger is removed
// (CR 603.3d). A target that has no +1/+1 counter by resolution loses
// nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b1f441ce-7619-4ab5-ab36-724ed0a76728",
		Name:         "Bloodcrazed Hoplite",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Heroic("Bloodcrazed Hoplite — a +1/+1 counter on it (heroic)", effects.ThisGetsAPlusOneCounter),
			effects.Targeting(effects.WheneverACounterIsPutOnThis(game.CounterPlusOne, "Bloodcrazed Hoplite — remove a +1/+1 counter from target creature an opponent controls",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						if c, ok := g.LookupCardForEffect(t.ID); ok && c.Counters[game.CounterPlusOne] > 0 {
							return effects.AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: -1}.Apply(ctx)
						}
					}
					return nil
				}), effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
