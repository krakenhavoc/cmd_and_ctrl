package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shield Broker — Creature — Octopus Advisor {3}{U}{U}, 3/4:
//
//	"When this creature enters, put a shield counter on target
//	 noncommander creature you don't control. You gain control of that
//	 creature for as long as it has a shield counter on it."
//
// The shield counter protects the stolen creature through the CR 122.1c
// built-ins (game/shield_counters.go, #2000), and the theft lasts while
// the counter does (game.WhilePinnedHasCounter, ADR 0109 §2): the first
// destruction or damage that spends the counter hands the creature back.
// A counter that couldn't be put on means the theft never starts (CR
// 611.2b).
func init() {
	Register(Spec{
		OracleID:     "481234e5-fbfd-46e8-8b9a-a4e708dd2195",
		Name:         "Shield Broker",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters(shieldBrokerLabel, resolveShieldBroker),
				TargetCreature("target noncommander creature you don't control", OpponentControls(), shieldBrokerNoncommander)),
		},
	})
}

const shieldBrokerLabel = "Shield Broker — put a shield counter on target noncommander creature you don't control and gain control of it while it has one"

// shieldBrokerNoncommander is "noncommander".
func shieldBrokerNoncommander(_ *game.Game, _ uuid.UUID, c game.Card) bool {
	return !c.IsCommander
}

func resolveShieldBroker(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	target := FirstLegalBattlefieldTarget(ctx)
	if target == uuid.Nil {
		return nil
	}
	if err := (AddCounter{Target: target, Kind: game.CounterShield, N: 1}).Apply(ctx); err != nil {
		return err
	}
	d, ok := DurationWhileItHasCounter(ctx, target, game.CounterShield)
	if !ok {
		return nil
	}
	return GainControl{Target: target, Controller: ctx.Controller(), Duration: d, Label: "Shield Broker — control while it has a shield counter"}.Apply(ctx)
}
