package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tar Pit Warrior — Creature — Cyclops Warrior {2}{B}, 3/4 (Visions, S82 older
// sets batch 792, #5631):
//
//	"When this creature becomes the target of a spell or ability, sacrifice
//	 it."
//
// Mirozel's "becomes the target" row (EventBecomesTarget on the source, one
// trigger for each targeting spell or ability, CR 115.3), sacrificing the
// Warrior as the trigger resolves, above the spell or ability that targets it.
// A Warrior that has left, or come back as a new object, is not sacrificed
// (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "05a7ca83-e820-433f-b9e9-151e817d3708",
		Name:         "Tar Pit Warrior",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesTarget, effects.Self, "Tar Pit Warrior — sacrifice it", effects.SacrificeThisIfStillOnBattlefield),
		},
	})
}
