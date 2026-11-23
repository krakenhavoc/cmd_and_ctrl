package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Barrier of Bones — Creature — Skeleton Wall {B}, 0/3 (Guilds of Ravnica, #5525):
//
//	"Defender
//	 When this creature enters, surveil 1."
//
// Defender is printed data; the enter trigger is Coin of Fate's surveil 1
// on the stack (CR 701.25a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4f25b125-1ded-4a34-9ed1-e6ce087ec48d",
		Name:            "Barrier of Bones",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Barrier of Bones — surveil 1", effects.Do(effects.Surveil{N: 1})),
		},
	})
}
