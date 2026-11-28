package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Village Elder — Creature — Human Druid {G}, 1/1 (Mirage, S82 OLD b808):
//
//	"{G}, {T}, Sacrifice a Forest: Regenerate target creature."
//
// Heartwood Giant's mana, tap and sacrifice-a-Forest cost around a
// regeneration shield (CR 701.19a) on the target, which does nothing if it has
// left as the ability resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "adea6cb4-1e74-4137-9c80-1dde004a6b26",
		Name:         "Village Elder",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{G}, {T}, Sacrifice a Forest: Regenerate target creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost: effects.Plus(effects.ManaCost("{G}"), effects.TapCost(),
				effects.SacrificeN(1, "a Forest", effects.Subtype("Forest"))),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.OldB808RegenerateTheTarget,
		}},
	})
}
