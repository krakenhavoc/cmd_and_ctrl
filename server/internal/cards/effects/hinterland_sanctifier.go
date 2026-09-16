package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hinterland Sanctifier — Creature — Rabbit Cleric {W}, 1/2 (EDHREC
// rank 2957):
//
//	"Whenever another creature you control enters, you gain 1 life."
//
// The Bloomburrow Soul Warden. b13AnotherCreatureYouControlEntered
// is the condition — any other creature entering under the
// controller's control, token or not, one trigger per creature — and
// each resolves for 1 life.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d812fc6d-b96d-4986-b171-9f3feee603dc",
		Name:         "Hinterland Sanctifier",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverAnotherCreatureEntersUnderYourControl("Hinterland Sanctifier — you gain 1 life", Do(GainLife{Amount: 1})),
		},
	})
}
