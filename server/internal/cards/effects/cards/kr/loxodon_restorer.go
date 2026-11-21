package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Loxodon Restorer — Creature — Elephant Cleric {4}{W}{W}, 3/4 (Guilds of Ravnica, #5525):
//
//	"Convoke
//	 When this creature enters, you gain 4 life."
//
// Convoke is the shared tap cost (CR 702.51a); the enter trigger is
// Obstinate Baloth's life gain.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "77eb2374-44a0-4383-b229-edc116c7d835",
		Name:         "Loxodon Restorer",
		Completeness: effects.CompletenessFull,
		TapCost:      effects.Convoke(),
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Loxodon Restorer — you gain 4 life", effects.Do(effects.GainLife{Amount: 4})),
		},
	})
}
