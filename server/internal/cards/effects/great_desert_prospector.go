package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Great Desert Prospector — Creature — Human Artificer {4}{W}, 3/2 (The
// Brothers' War, #5482):
//
//	"When this creature enters, create a tapped Powerstone token for
//	 each other creature you control."
//
// The count is read as the trigger resolves (Eagle of the Great
// Shelf's other-creature count), and the tokens are one creation.
// With no other creature, nothing is made.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "37278877-e7d8-47b6-a65f-37c7f85e3a0a",
		Name:         "Great Desert Prospector",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Great Desert Prospector — create a tapped Powerstone token for each other creature you control",
				greatDesertProspectorTokens),
		},
	})
}

func greatDesertProspectorTokens(g *game.Game, item *game.StackItem) error {
	n := 0
	for _, id := range permanentsControlledByMatching(g, item.Controller, Creature()) {
		if id != item.SourceCardID {
			n++
		}
	}
	if n == 0 {
		return nil
	}
	return CreateTokenAdvanced{Spec: Token(PowerstoneToken()).EntersTapped(), N: n}.Apply(NewContext(g, item))
}
