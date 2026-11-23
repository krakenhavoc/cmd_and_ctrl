package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Marauding Blight-Priest — Creature — Vampire Cleric {2}{B}, 3/2
// (EDHREC rank 1113):
//
//	"Whenever you gain life, each opponent loses 1 life."
//
// The lifegain deck's table-wide drain. Sanguine Bond's trigger with
// a fixed 1 and no target: every EventChangeLife with a positive
// delta on the controller — lifelink included — fires it once, and
// "gain 5 life" is one event, so one trigger, as printed. Life LOSS,
// not damage, so no prevention or doubling applies.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "814b87fe-2a75-4ff2-8637-7e69e3fb285b",
		Name:         "Marauding Blight-Priest",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouGainLife("Marauding Blight-Priest — each opponent loses 1 life", effects.Do(effects.EachOpponentLosesLife{N: 1})),
		},
	})
}
