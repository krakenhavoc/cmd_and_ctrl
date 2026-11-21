package testsso

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// spree_registry_test.go — the boot-time guards ADR 0065's
// 2026-09-23 amendment adds for a mode's Cost (CR 702.172a). Each
// panic pins a card-file mistake the four proof cards
// (three_steps_ahead.go, explosive_derailment.go,
// insatiable_avarice.go, caught_in_the_crossfire.go) never make —
// and their successful registration at package init is this file's
// positive case, so it is not repeated here.

func TestRegisterUnparseableModeCostPanics(t *testing.T) {
	effectstest.MustPanic(t, "unparseable Cost", func() {
		effects.Register(effects.Spec{
			OracleID: "spree-registry-test-unparseable",
			Name:     "Bad Cost Spree",
			Modes: effects.Spree(
				effects.SpreeMode("Bullet one.", "not a cost"),
				effects.SpreeMode("Bullet two.", "{1}"),
			),
		})
	})
}

// CR 702.172a prices a spell's own modes. No printed activated
// ability charges more for choosing one of its modes, so the field
// has nowhere to go on that owner yet.
func TestRegisterModeCostOnActivatedAbilityPanics(t *testing.T) {
	effectstest.MustPanic(t, "Spree", func() {
		effects.Register(effects.Spec{
			OracleID: "spree-registry-test-activated",
			Name:     "Bad Cost Activated",
			Activated: []effects.ActivatedAbility{{
				Label: "Activated bullets",
				Cost:  effects.TapCost(),
				Modes: &game.ModeSpec{
					Prompt: "Choose one", Min: 1, Max: 1,
					Options: []game.ModeOption{
						{Label: "bullet one", Cost: "{1}",
							Effect: func(*game.Game, *game.StackItem, int) error { return nil }},
					},
				},
			}},
		})
	})
}

// Same reasoning, the trigger owner.
func TestRegisterModeCostOnTriggerPanics(t *testing.T) {
	effectstest.MustPanic(t, "Spree", func() {
		effects.Register(effects.Spec{
			OracleID: "spree-registry-test-trigger",
			Name:     "Bad Cost Trigger",
			Triggered: []game.TriggeredAbility{{
				Watches: []game.EventKind{game.EventETB},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return ev.CardID == source.InstanceID
				},
				Modes: &game.ModeSpec{
					Prompt: "Choose one", Min: 1, Max: 1,
					Options: []game.ModeOption{
						{Label: "bullet one", Cost: "{1}",
							Effect: func(*game.Game, *game.StackItem, int) error { return nil }},
					},
				},
			}},
		})
	})
}
