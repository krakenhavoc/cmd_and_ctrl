package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ray of Command — Instant {3}{U} (Commander 2011, #5574):
//
//	"Untap target creature an opponent controls and gain control of it until
//	 end of turn. That creature gains haste until end of turn. When you lose
//	 control of the creature, tap it."
//
// Take for a Ride's steal (StdB31Threaten: control until end of turn, untap,
// haste). The last sentence is a delayed trigger (CR 603.7) watching
// EventControlChanged for that creature leaving the spell's controller, whether
// the until-end-of-turn duration ended in the cleanup step or another effect
// took it first. It taps the creature, which is the old owner's again by then.
// A creature that has left the battlefield is not tapped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "adf41d29-403d-4936-9689-3148d103e700",
		Name:         "Ray of Command",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := effects.FirstLegalTarget(ctx)
			if !ok {
				return nil
			}
			if err := effects.StdB31Threaten(ctx, t, "Ray of Command"); err != nil {
				return err
			}
			return effects.DelayedOnEvent{
				Label:     "Ray of Command — tap the creature",
				On:        []game.EventKind{game.EventControlChanged},
				Condition: rayOfCommandLostControl,
				Cards:     []uuid.UUID{t.ID},
				Body:      rayOfCommandTapBody,
				// The theft ends in the cleanup step, which is still this turn.
				ThroughCleanup: true,
			}.Apply(ctx)
		},
	})
}

// rayOfCommandLostControl is "when you lose control of the creature": the
// creature changed controller away from the trigger's controller. An on-disk key.
var rayOfCommandLostControl = game.DelayedCondition("ray-of-command/lost-control-of-the-creature",
	func(ev game.Event, dt *game.DelayedTrigger, _ *game.Game, _ game.EffectParams) bool {
		return ev.Kind == game.EventControlChanged && ev.Target == dt.Controller && len(dt.Cards) > 0 && ev.CardID == dt.Cards[0]
	})

// rayOfCommandTapBody taps the creature. An on-disk key.
var rayOfCommandTapBody = game.SimpleDelayedBody("ray-of-command/tap-the-creature",
	func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		for _, t := range item.Targets {
			if t.Kind != game.TargetCard {
				continue
			}
			if err := (effects.TapTarget{Target: t.ID}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})
