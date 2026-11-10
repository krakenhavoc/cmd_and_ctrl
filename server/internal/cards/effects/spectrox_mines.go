package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spectrox Mines — Plane — Androzani Minor (Doctor Who, #4870):
//
//	"When you planeswalk to Spectrox Mines and at the beginning of your
//	 upkeep, you lose 3 life and create a Treasure token.
//	 Whenever chaos ensues, create a Food token and a 2/2 black Human
//	 Rogue creature token."
//
// A plane (ADR 0216 §6), controlled by the planar controller (CR 901.6).
// The first ability is one ability with two trigger conditions. The chaos
// ability (CR 311.7) creates both tokens for the planar controller.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4f387b4c-3e4a-419d-9dbe-95c5f797747e",
		Name:         "Spectrox Mines",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenYouPlaneswalkToThisAndAtYourUpkeep("Spectrox Mines — you lose 3 life and create a Treasure token",
				Do(planesB2LoseLife{Amount: 3}, CreateToken{Template: TreasureToken(), N: 1})),
			WheneverChaosEnsues("Spectrox Mines — create a Food token and a 2/2 black Human Rogue creature token",
				Do(CreateToken{Template: FoodToken(), N: 1}, CreateToken{Template: TokenCard("2/2 black Human Rogue"), N: 1})),
		},
	})
}
