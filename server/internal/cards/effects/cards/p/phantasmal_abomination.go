package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Phantasmal Abomination — {1}{U}{U} Creature — Illusion 5/5 (Rise of the Eldrazi, S82 OLD #5581):
//
//	Defender
//	When this creature becomes the target of a spell or ability, sacrifice it.
//
// Skulking Ghost's trigger: it fires on the targeting (CR 603.2), goes on the stack above the targeting spell or ability, and sacrifices the Abomination if it is still on the battlefield.

func init() {
	effects.Register(effects.Spec{
		OracleID:        "f4925e05-1332-4466-b2ea-9f6f4ccf7db4",
		Name:            "Phantasmal Abomination",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesTarget, effects.Self, "Phantasmal Abomination — sacrifice it", effects.SacrificeThisIfStillOnBattlefield),
		},
	})
}
