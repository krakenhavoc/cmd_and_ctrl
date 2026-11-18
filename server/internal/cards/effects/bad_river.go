package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bad River — Land:
//
//	"This land enters tapped.
//	 {T}, Sacrifice this land: Search your library for an Island or Swamp
//	 card, put it onto the battlefield, then shuffle."
//
// The fetch is the shared fetchDual search without the life payment;
// the land it finds enters untapped, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "259fc423-d078-4668-9e95-8c1a0d2e49ba",
		Name:         "Bad River",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		Activated: []ActivatedAbility{{
			Label:   "{T}, Sacrifice this land: Search your library for an Island or Swamp card, put it onto the battlefield, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Lands: 1, LandsUntapped: 1},
			Cost:    Plus(TapCost(), SacrificeThis()),
			Effect:  fetchDual("island", "swamp"),
		}},
	})
}
