package testsu

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// x_bound_activated_test.go — #1723's registration-time half: an
// X-bound target clause (ManaValueAtMostX / ManaValueEqualsX) is
// legal on an activated ability's clause exactly when that ability's
// OWN cost announces an X (CR 602.2b), same as it always was on a
// trigger's clause (never legal — a trigger announces no X at all).

// TestRegisterXBoundActivatedClauseNeedsCostX pins the refusal: an
// activated ability with an X-bound target but a cost that never
// announces X (no {X}, no variable sacrifice/tap count) is exactly
// the "a trigger announces no X" mistake one owner over, and Register
// still catches it.
func TestRegisterXBoundActivatedClauseNeedsCostX(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Register did not panic for an X-bound activated clause with no {X} in the cost")
		}
	}()
	effects.Register(effects.Spec{
		OracleID: "test-x-bound-activated-no-x",
		Name:     "No X Test Card",
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}: does nothing",
			Cost:    effects.ManaCost("{1}"),
			Targets: effects.TargetAny().WithManaValueEqualsX(),
			Effect:  func(_ *game.Game, _ *game.StackItem) error { return nil },
		}},
	})
}

// TestRegisterXBoundActivatedClauseWithCostXIsFine is the positive
// case Lazav, the Multifarious needs: an activated ability whose OWN
// cost has {X} may bind a target clause to that X.
func TestRegisterXBoundActivatedClauseWithCostXIsFine(t *testing.T) {
	effectstest.RegisterForTest(t, effects.Spec{
		OracleID: "test-x-bound-activated-with-x",
		Name:     "X Test Card",
		Activated: []effects.ActivatedAbility{{
			Label:   "{X}: does nothing",
			Cost:    effects.ManaCost("{X}"),
			Targets: effects.TargetAny().WithManaValueEqualsX(),
			Effect:  func(_ *game.Game, _ *game.StackItem) error { return nil },
		}},
	})
}

// TestRegisterXBoundTriggerStillPanics: unlike an activated ability, a
// trigger never announces an X, whatever the card's other abilities'
// costs look like — #1723 only widened the activated-ability case.
func TestRegisterXBoundTriggerStillPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Register did not panic for an X-bound trigger clause")
		}
	}()
	effects.Register(effects.Spec{
		OracleID: "test-x-bound-trigger",
		Name:     "Trigger X Test Card",
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: func(game.Event, *game.Card, game.Characteristic, *game.Game) bool { return true },
			Targets:   effects.TargetAny().WithManaValueEqualsX(),
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "does nothing")
			},
		}},
	})
}

// TestRegisterBothManaValueFlagsPanics: "or less" and "exactly" are
// different clauses, never one card's.
func TestRegisterBothManaValueFlagsPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Register did not panic for a clause with both ManaValueAtMostX and ManaValueEqualsX")
		}
	}()
	effects.Register(effects.Spec{
		OracleID: "test-x-bound-both-flags",
		Name:     "Both Flags Test Card",
		Targets:  effects.TargetAny().WithManaValueAtMostX().WithManaValueEqualsX(),
	})
}
