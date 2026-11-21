package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mistrise Village — Land:
//
//	"This land enters tapped unless you control a Mountain or a Forest.
//	 {T}: Add {U}.
//	 {U}, {T}: The next spell you cast this turn can't be countered."
//
// The checkland condition (youControlLandTyped, check_lands.go) on a
// mono-blue land, and Insist's promise as an activated ability (ADR
// 0106 §4 decision 3, #1806): the next spell its controller casts this
// turn, of any kind, spends the promise as it becomes cast (CR 601.2i)
// and can't be countered while it is on the stack. Activated with a
// spell already on the stack, it covers the NEXT one cast, not that
// one. An unspent promise ends at cleanup (CR 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "339f5334-b65a-445a-a016-20e997e0b4bb",
		Name:         "Mistrise Village",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTappedUnless(effects.YouControlLandTyped("mountain", "forest"))},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{U}",
			Label:    "Add {U}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}, {T}: The next spell you cast this turn can't be countered.",
			Purpose: game.Purpose{Answers: game.AnswerProtect},
			Cost:    effects.Plus(effects.ManaCost("{U}"), effects.TapCost()),
			Effect: effects.Do(effects.GrantCounterShield{From: "Mistrise Village", Grant: effects.NextSpellYouCastCantBeCountered(
				"The next spell you cast this turn can't be countered.", game.PermissionFilter{})}),
		}},
	})
}
