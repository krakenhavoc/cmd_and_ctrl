package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ancestral Statue — Artifact Creature — Golem {4}, 3/4 (Jumpstart, #5510):
//
//	"When this creature enters, return a nonland permanent you control to
//	 its owner's hand."
//
// Roaring Primadox's mandatory pick as the trigger resolves, over the
// nonland permanents the controller has then, the Statue itself included
// (it is the only legal pick when it is the only nonland permanent).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "da69fbf7-c894-41dd-85ec-217de23a362f",
		Name:         "Ancestral Statue",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Ancestral Statue — return a nonland permanent you control to its owner's hand",
				effects.Do(effects.ReturnOneYouControl{
					Match:    func(c game.Card) bool { return !c.IsLand() },
					Question: "Ancestral Statue — return a nonland permanent you control to its owner's hand",
				})),
		},
	})
}
