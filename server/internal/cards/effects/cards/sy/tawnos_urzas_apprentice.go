package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tawnos, Urza's Apprentice — Legendary Creature — Human Artificer {U}{R},
// 1/3 (Commander 2018, #5526):
//
//	"Haste
//	 {U}{R}, {T}: Copy target activated or triggered ability you control
//	 from an artifact source. You may choose new targets for the copy.
//	 (Mana abilities can't be targeted.)"
//
// Weaver of Harmony's ability with an artifact source: the clause is
// AbilityOnStack over an ability you control whose source is an artifact
// (AbilityFromSource(Artifact()), read off the source as it is now), and
// the copy is Lithoform Engine's copyTargetedAbility, with the new-targets
// offer. A mana ability never uses the stack (CR 605.3b), so there is none
// to offer. Haste lets Tawnos tap the turn he arrives.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e835da40-f484-4d8b-98de-595cd23b51bc",
		Name:            "Tawnos, Urza's Apprentice",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Activated: []effects.ActivatedAbility{{
			Label: "{U}{R}, {T}: Copy target activated or triggered ability you control from an artifact source",
			Cost:  effects.Plus(effects.ManaCost("{U}{R}"), effects.TapCost()),
			Targets: effects.AbilityOnStack("target activated or triggered ability you control from an artifact source",
				effects.AnAbilityYouControl(), effects.AbilityFromSource(effects.Artifact())),
			Effect: effects.CopyTargetedAbility,
		}},
	})
}
