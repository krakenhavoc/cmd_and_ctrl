package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rescuer Chwinga — Creature — Elemental Spirit {1}{W}, 2/2 (Commander
// Legends: Battle for Baldur's Gate, #5948):
//
//	"Flash
//	 Natural Shelter — When this creature enters, you may return another
//	 permanent you control to its owner's hand."
//
// Nothing targets: the permanent is chosen as the trigger resolves (First
// Responder's "another creature you control" pick, over any permanent), zero
// is the "you may" decline, and the Chwinga itself is never offered.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "be1f0d72-14b5-41be-a37a-7abeb52ec0e5",
		Name:            "Rescuer Chwinga",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Rescuer Chwinga — Natural Shelter: you may return another permanent you control to its owner's hand",
				func(g *game.Game, item *game.StackItem) error {
					return effects.ChooseAnotherPermanentYouControl(g, item,
						"Rescuer Chwinga — return another permanent you control to its owner's hand?", effects.GapB225ReturnPicked)
				}),
		},
	})
}
