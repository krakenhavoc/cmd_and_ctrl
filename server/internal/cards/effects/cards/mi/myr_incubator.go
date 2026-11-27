package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Myr Incubator — Artifact {6} (Mirrodin, #5606):
//
//	"{6}, {T}, Sacrifice this artifact: Search your library for any number of
//	 artifact cards, exile them, then create that many 1/1 colorless Myr
//	 artifact creature tokens. Then shuffle."
//
// Bilbo's "any number of" search (a real choice from none to every artifact
// card, the zero floor of CR 701.23b) to exile, and a tail that makes one Myr
// token for each card it exiled. The Incubator is sacrificed as the cost, so
// it is never among its own finds.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d0a493af-551e-4475-af5d-ec4e080728cb",
		Name:         "Myr Incubator",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{6}, {T}, Sacrifice this artifact: Search your library for any number of artifact cards, exile them, then create that many 1/1 colorless Myr artifact creature tokens. Then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{6}"), effects.TapCost(), effects.SacrificeThis()),
			Effect:  effects.OldB607MyrIncubatorSearch,
		}},
	})
}
