package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Maze Abomination — Creature — Elemental {5}{B}, 4/5:
//
//	"Deathtouch (Any amount of damage this deals to a creature is
//	 enough to destroy it.)
//	 Multicolored creatures you control have deathtouch."
//
// The printed deathtouch plus a layer-6 grant (CR 613.1f) of deathtouch to
// every creature its controller controls that is two or more colours
// (CR 105.2b), read through the layers like Rienne's multicolored anthem.
// A monocolored or colourless creature gets nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "eaa72823-413e-474e-a926-aebb0a78ff31",
		Name:            "Maze Abomination",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Static:          []game.StaticAbility{effects.DgmMulticoloredCreaturesHave("deathtouch")},
	})
}
