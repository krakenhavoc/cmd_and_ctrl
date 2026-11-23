package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sprouting Thrinax — Creature — Lizard {B}{R}{G}, 3/3 (Shards of Alara,
// #5588):
//
//	"When this creature dies, create three 1/1 green Saproling creature
//	 tokens."
//
// Tukatongue Thallid's dies trigger making three: a leaves-the-battlefield
// trigger that looks back in time (CR 603.10a), controlled by the
// Thrinax's controller as it died.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ea74e89e-0724-4744-a6c9-e35c415624fb",
		Name:         "Sprouting Thrinax",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.WhenThisDies("Sprouting Thrinax — create three 1/1 green Saproling creature tokens",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 green Saproling"), N: 3})), game.Purpose{Tokens: 3}),
		},
	})
}
