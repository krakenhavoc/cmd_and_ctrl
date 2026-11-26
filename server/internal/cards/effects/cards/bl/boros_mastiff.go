package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Boros Mastiff — Creature — Dog {1}{W}, 2/2 (Dragon's Maze, S82 OLD b406,
// #5564):
//
//	"Battalion — Whenever this creature and at least two other creatures
//	 attack, this creature gains lifelink until end of turn."
//
// Ordruun Veteran's battalion trigger (GtcBattalion) with lifelink in place of
// double strike.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bcbaef07-1d5b-457b-b373-77d3a2884b39",
		Name:         "Boros Mastiff",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.GtcBattalion("Boros Mastiff — it gains lifelink until end of turn (battalion)",
				effects.ThisGainsKeywordUntilEndOfTurn("lifelink", "Boros Mastiff — lifelink until end of turn")),
		},
	})
}
