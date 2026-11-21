package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Echo, Perceptive Prodigy — {2}{U} Legendary Creature — Human Hero 1/4:
//
//	"Vigilance
//	 {1}, {T}: Copy target activated or triggered ability you control from a
//	 creature source. You may choose new targets for the copy. (Mana
//	 abilities can't be targeted.)"
//
// Weaver of Harmony's shape with a different source clause: the target is an
// ability on the stack (AbilityOnStack) that its source, read as it is now,
// is a creature (AbilityFromSource). A mana ability never uses the stack
// (CR 605.3b), so the parenthetical needs no code.
//
// No simplification. Marvel Super Heroes (#4455).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1091c83a-a4b9-457f-a69a-9a44a7e10759",
		Name:            "Echo, Perceptive Prodigy",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated: []effects.ActivatedAbility{{
			Label: "{1}, {T}: Copy target activated or triggered ability you control from a creature source",
			Cost:  effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			Targets: effects.AbilityOnStack("target activated or triggered ability you control from a creature source",
				effects.AnAbilityYouControl(), effects.AbilityFromSource(effects.Creature())),
			Effect: effects.CopyTargetedAbility,
		}},
	})
}
