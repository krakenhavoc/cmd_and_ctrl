package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Woodwraith Strangler — Creature — Plant Zombie {2}{B}{G}, 2/2 (Ravnica: City of
// Guilds):
//
//	"Exile a creature card from your graveyard: Regenerate this creature."
//
// Moorland Haunt's cost shape (ExileFromGraveyard, #1297): the activator names
// a creature card in their own graveyard at announce (CR 602.2b), which is
// exiled as the cost is paid (CR 601.2h), and the ability makes one
// regeneration shield on the Strangler (CR 701.19a). The Strangler itself is
// never a legal pick. A source that has left the battlefield gets nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c72ab66b-2011-449c-93a6-1808669653f0",
		Name:         "Woodwraith Strangler",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Exile a creature card from your graveyard: Regenerate this creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.ExileFromGraveyard(1, "a creature card", effects.IsCreatureCard),
			Effect:  effects.RegenerateThisSource,
		}},
	})
}
