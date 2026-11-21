package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lumbering Megasloth — Creature — Sloth Mutant {10}{G}{G}, 8/8 (Fallout,
// #5948):
//
//	"This spell costs {1} less to cast for each counter among players and
//	 permanents.
//	 Trample
//	 This creature enters tapped."
//
// A self cost reduction (CostsLessEach) over every counter of every kind on
// every player and every permanent on the battlefield, whoever controls it
// (gapB63CountersAmongPlayersAndPermanents), read as the cost is determined
// (CR 601.2f); the engine takes the reduction off the generic part only
// (CR 601.2f, CR 118.7), so it never goes below {G}{G}. Enters tapped is a
// replacement (SelfEntersTapped).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6492c381-0fa1-4d43-bf72-d8956d8a34b6",
		Name:            "Lumbering Megasloth",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessEach(effects.GapB63CountersAmongPlayersAndPermanents,
				"This spell costs {1} less to cast for each counter among players and permanents."),
		},
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
	})
}
