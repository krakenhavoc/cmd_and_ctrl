package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Prosperous Innkeeper — Creature — Halfling Citizen {1}{G}, 1/1
// (EDHREC rank 1498):
//
//	"When this creature enters, create a Treasure token. (It's an
//	 artifact with "{T}, Sacrifice this token: Add one mana of any
//	 color.")
//	 Whenever another creature you control enters, you gain 1 life."
//
// A two-drop that ramps and then drips life. Two ordinary triggers:
// the ETB makes a real Treasure; the alliance half is one trigger
// per creature entering, and the gain goes through the life
// pipeline so every "whenever you gain life" payoff sees it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "da785227-cf8a-4d44-9e7c-fc909ea868f2",
		Name:         "Prosperous Innkeeper",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Prosperous Innkeeper — create a Treasure", Do(CreateToken{Template: TreasureToken(), N: 1})),
			WheneverAnotherCreatureEntersUnderYourControl("Prosperous Innkeeper — you gain 1 life", Do(GainLife{Amount: 1})),
		},
	})
}
