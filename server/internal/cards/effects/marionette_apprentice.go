package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Marionette Apprentice — Creature — Human Artificer {1}{B}:
//
//	"Fabricate 1 (When this creature enters, put a +1/+1 counter on it
//	 or create a 1/1 colorless Servo artifact creature token.)
//	 Whenever another creature or artifact you control is put into a
//	 graveyard from the battlefield, each opponent loses 1 life."
//
// Fabricate is the shared keyword constructor. The drain reads the
// departure, so it counts what the permanent was on the battlefield
// (a token Servo is both a creature and an artifact and drains once),
// and it fires for each permanent in a simultaneous wipe.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "726d9d2c-736a-4852-9938-a0f50d8fd89f",
		Name:         "Marionette Apprentice",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Fabricate("Marionette Apprentice", 1),
			On(game.EventLTB, AnotherCreatureOrArtifactYouControlWasPutIntoAGraveyard,
				"Marionette Apprentice — each opponent loses 1 life",
				func(g *game.Game, item *game.StackItem) error {
					return eachOpponentLosesLife(g, item, 1)
				}),
		},
	})
}
