package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tel-Jilad Lifebreather — Creature — Troll Shaman {4}{G}, 3/2 (Fifth Dawn, #5604):
//
//	"{G}, {T}, Sacrifice a Forest: Regenerate target creature."
//
// Goblin Chirurgeon's regeneration of any target creature, paid with {G},
// {T} and a Forest (a land with the Forest subtype, basic or not).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1ef49064-eb2a-49ae-917e-3d441ba14bf9",
		Name:         "Tel-Jilad Lifebreather",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{G}, {T}, Sacrifice a Forest: Regenerate target creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.Plus(effects.ManaCost("{G}"), effects.TapCost(), effects.SacrificeN(1, "a Forest", effects.Subtype("Forest"))),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.RegenerateTheTargetPermanent,
		}},
	})
}
