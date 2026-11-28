package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Zombie Scavengers — Creature — Zombie {2}{B}, 3/1 (Weatherlight, S82 OLD b773):
//
//	"Exile the top creature card of your graveyard: Regenerate this creature."
//
// The cost is the topmost creature card of the activator's graveyard: "the
// top creature card" is the first creature card counting from the top
// (CR 404.2, #6194). The regeneration shield is the creature's own
// (CR 701.19a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8c5298da-6759-444e-8870-6c96d226c940",
		Name:         "Zombie Scavengers",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Exile the top creature card of your graveyard: Regenerate this creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.ExileTopFromGraveyard(1, "the top creature card of your graveyard", effects.MatchCreature),
			Effect:  effects.RegenerateThisSource,
		}},
	})
}
