package testst

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestRegisterRefusesMalformedTapOthersCostsOnBothAbilityKinds(t *testing.T) {
	validFilter := effects.TargetPermanent("an untapped creature you control", effects.Creature())
	cases := []struct {
		name string
		cost *game.TapOthersCost
		want string
	}{
		{name: "zero count", cost: &game.TapOthersCost{Filter: validFilter, Label: "a creature"}, want: "positive count"},
		{name: "missing filter", cost: &game.TapOthersCost{Count: 1, Label: "a creature"}, want: "clause and label"},
		{name: "missing label", cost: &game.TapOthersCost{Count: 1, Filter: validFilter}, want: "clause and label"},
		{name: "same permanent twice", cost: &game.TapOthersCost{
			Count:  2,
			Filter: &game.TargetSpec{AllowSame: true},
			Label:  "two permanents",
		}, want: "AllowSame"},
		{name: "players", cost: &game.TapOthersCost{
			Count:  1,
			Filter: &game.TargetSpec{Players: true},
			Label:  "a player",
		}, want: "matches permanents only"},
	}

	for _, tc := range cases {
		for _, owner := range []string{"activated", "mana"} {
			t.Run(tc.name+"/"+owner, func(t *testing.T) {
				spec := effects.Spec{OracleID: "test-tap-others-register-" + tc.name + "-" + owner, Name: "Bad Tap Cost"}
				if owner == "activated" {
					spec.Activated = []effects.ActivatedAbility{{Cost: game.AbilityCost{TapOthers: tc.cost}}}
				} else {
					spec.ManaAbilities = []effects.ManaAbility{{Cost: effects.ManaAbilityCost{TapOthers: tc.cost}, Produced: "{G}"}}
				}
				if msg := effectstest.RegisterPanics(spec); !strings.Contains(msg, tc.want) {
					t.Fatalf("Register panic = %q, want it to contain %q", msg, tc.want)
				}
			})
		}
	}
}

func TestRegisterAllowsTapXOnlyOnAnOrdinaryActivatedAbility(t *testing.T) {
	variable := &game.TapOthersCost{
		Filter: &game.TargetSpec{CountFromX: true},
		Label:  "X permanents",
	}
	// No panic: a CR 602 activation has an X announcement and a stack
	// item to carry it.
	effects.CheckTapOthersClause("Variable Tapper", "ability 0", variable, true)

	spec := effects.Spec{
		OracleID: "test-variable-tap-mana",
		Name:     "Bad Variable Mana Tapper",
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{TapOthers: variable},
			Produced: "{G}",
		}},
	}
	if msg := effectstest.RegisterPanics(spec); !strings.Contains(msg, "mana ability has no X announcement") {
		t.Fatalf("Register panic = %q, want the mana-ability X refusal", msg)
	}
}
