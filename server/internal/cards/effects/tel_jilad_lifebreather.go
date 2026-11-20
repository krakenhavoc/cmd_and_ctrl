package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tel-Jilad Lifebreather — Creature — Troll Shaman {4}{G}, 3/2 (Fifth Dawn, #5604):
//
//	"{G}, {T}, Sacrifice a Forest: Regenerate target creature."
//
// Goblin Chirurgeon's regeneration of any target creature, paid with {G},
// {T} and a Forest (a land with the Forest subtype, basic or not).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1ef49064-eb2a-49ae-917e-3d441ba14bf9",
		Name:         "Tel-Jilad Lifebreather",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{G}, {T}, Sacrifice a Forest: Regenerate target creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    Plus(ManaCost("{G}"), TapCost(), SacrificeN(1, "a Forest", Subtype("Forest"))),
			Targets: TargetCreature("target creature"),
			Effect:  regenerateTheTargetPermanent,
		}},
	})
}
