package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Yavimaya Hollow — Legendary Land:
//
//	"{T}: Add {C}.
//	 {G}, {T}: Regenerate target creature."
//
// A colorless mana ability and a regeneration ability on one land; the
// two share {T}, so tapping for mana spends the regeneration and the
// other way round. The second is an ordinary activated ability built
// like Asceticism's — {G} plus {T}, target creature, the shared
// regenerateTheTargetPermanent effect — so it uses the stack and can be
// answered, and the shield it makes replaces the NEXT destruction of
// that creature this turn (CR 701.19a). Legendary is printed type-line
// data the engine already carries.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "53d6113d-acdb-4754-9641-f7991a96c7b9",
		Name:         "Yavimaya Hollow",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{G}, {T}: Regenerate target creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.Plus(effects.ManaCost("{G}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.RegenerateTheTargetPermanent,
		}},
	})
}
