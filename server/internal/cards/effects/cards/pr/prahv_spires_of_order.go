package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Prahv, Spires of Order — Land:
//
//	"{T}: Add {C}.
//	 {4}{W}{U}, {T}: Prevent all damage a source of your choice would deal this turn."
//
// ADR 0108 §7 (#1904): a shield against a source chosen as the ability
// resolves (CR 609.7a), preventing all of that source's damage, to
// anything, for the rest of the turn.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "37ff5ba6-0763-4c73-85bf-66856e67b8f3",
		Name:         "Prahv, Spires of Order",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{effects.SourceShieldRow(
			"{4}{W}{U}, {T}: Prevent all damage a source of your choice would deal this turn.",
			effects.Plus(effects.ManaCost("{4}{W}{U}"), effects.TapCost()), nil,
			effects.PreventDamageFromChosenSource(effects.ShieldAnything))},
	})
}
