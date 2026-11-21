package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Collation Trimatrix — Artifact Creature — Borg {B}, 1/1 (Star Trek):
//
//	Tap three untapped artifact creatures you control: You draw a card and
//	lose 1 life.
//
// The cost is a TapOthers component, not the {T} symbol, so the Trimatrix
// may tap itself as one of the three and a creature that arrived this turn
// may pay (CR 302.6). It is not a mana ability, so it uses the stack.
//
// No simplification.
func init() {
	const tapThree = "three untapped artifact creatures you control"
	effects.Register(effects.Spec{
		OracleID:     "4042ab14-c579-4d82-a642-576b88508ec7",
		Name:         "Collation Trimatrix",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label: "Tap three untapped artifact creatures you control: You draw a card and lose 1 life.",
			Cost: game.AbilityCost{TapOthers: &game.TapOthersCost{
				Count:  3,
				Filter: effects.TargetPermanent(tapThree, effects.Artifact(), effects.Creature()),
				Label:  tapThree,
			}},
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
			Effect:  effects.B36DrawAndLoseOne,
		}},
	})
}
