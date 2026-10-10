package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "339f5334-b65a-445a-a016-20e997e0b4bb",
		Name:         "Mistrise Village",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTappedUnless(youControlLandTyped("mountain", "forest"))},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{U}",
			Label:    "Add {U}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{U}, {T}: The next spell you cast this turn can't be countered.",
			Purpose: game.Purpose{Answers: game.AnswerProtect},
			Cost:    Plus(ManaCost("{U}"), TapCost()),
			Effect: Do(GrantCounterShield{From: "Mistrise Village", Grant: NextSpellYouCastCantBeCountered(
				"The next spell you cast this turn can't be countered.", game.PermissionFilter{})}),
		}},
	})
}
