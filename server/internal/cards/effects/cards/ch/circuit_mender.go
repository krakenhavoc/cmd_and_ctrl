package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Circuit Mender — Artifact Creature — Insect {3}, 2/3 (EDHREC rank
// 2204):
//
//	"When this creature enters, you gain 2 life.
//	 When this creature leaves the battlefield, draw a card."
//
// The colourless value body. Two triggers: a mandatory ETB lifegain
// and a leaves-the-battlefield draw that fires on ANY exit — dying,
// exile, bounce, a flicker — which is the Slithermuse shape, not the
// dies-only cardDied gate.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1665ca9f-176d-40f1-a4e9-42da4f1236e9",
		Name:         "Circuit Mender",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Circuit Mender — gain 2 life", effects.Do(effects.GainLife{Amount: 2})),
			effects.On(game.EventLTB, effects.Self, "Circuit Mender — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
