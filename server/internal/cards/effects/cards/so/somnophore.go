package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Somnophore — Creature — Illusion 2/2 {2}{U}{U} (Urza's Saga, S82 OLD b723):
//
//	Flying
//	Whenever this creature deals damage to a player, tap target creature that player controls. That creature doesn't untap during its controller's untap step for as long as this creature remains on the battlefield.
//
// Aberrant's "target … that player controls" on a damage trigger (any damage, combat or not). The tap and the hold are one effect: the hold lasts while this is the same permanent on the battlefield (CR 611.2b), and a Somnophore that left before the ability resolves taps the target and holds nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "86541bba-7db8-4893-addf-8a4a42168ce4",
		Name:            "Somnophore",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventDealDamage},
			AppliesTo: effects.ThisDealtDamageToAPlayer,
			Key:       "Somnophore — tap target creature that player controls; it doesn't untap while Somnophore remains on the battlefield",
			Targets:   effects.ControlledByThatPlayer(effects.TargetCreature("target creature that player controls")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				targets := effects.HoldTargetIDs(ctx)
				if err := effects.TapEach(ctx, targets); err != nil {
					return err
				}
				d, ok := effects.DurationWhileSourceRemains(ctx, ctx.Source())
				if !ok {
					return nil
				}
				return effects.DoesntUntapWhile{Targets: targets, Duration: d}.Apply(ctx)
			},
		}},
	})
}
