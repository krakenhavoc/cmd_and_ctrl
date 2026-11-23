package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kor Haven — Legendary Land:
//
//	"{T}: Add {C}.
//	 {1}{W}, {T}: Prevent all combat damage that would be dealt by target attacking creature this turn."
//
// ADR 0108 §7 (#1904, Delivery PR 7b): the shield's source is the
// targeted attacker, pinned as the ability resolves (CR 400.7); its
// combat damage to anything is prevented for the rest of the turn.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "276cece9-f9f2-46e6-ae76-daddaa2fb9ab",
		Name:         "Kor Haven",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{effects.ShieldAgainstTargetsRow(
			"{1}{W}, {T}: Prevent all combat damage that would be dealt by target attacking creature this turn.",
			effects.Plus(effects.ManaCost("{1}{W}"), effects.TapCost()),
			effects.TargetCreature("target attacking creature", effects.AttackingCreature()), true)},
	})
}
