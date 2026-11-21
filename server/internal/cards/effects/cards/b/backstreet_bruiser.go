package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Backstreet Bruiser — Creature — Octopus Rogue {1}{U}, 3/3 (Streets of
// New Capenna):
//
//	"Defender
//	 As long as there are two or more counters among creatures you
//	 control, this creature can attack as though it didn't have defender."
//
// The conditional defender waiver (CR 702.3b), Bristlepack Sentry's shape:
// the keyword stays and the waiver is read live at declaration. The count
// is every counter of any kind on creatures the controller controls,
// this one included, summed rather than counted by creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "354d9db1-18a6-4701-be3c-c4ded900e2ff",
		Name:            "Backstreet Bruiser",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Static: []game.StaticAbility{effects.AvB18CanAttackAsThoughNoDefenderWhile(
			"As long as there are two or more counters among creatures you control, this creature can attack as though it didn't have defender.",
			func(g *game.Game, source *game.Card) bool {
				return effects.CountersAmongCreaturesControlledBy(g, source.Controller) >= 2
			})},
	})
}
