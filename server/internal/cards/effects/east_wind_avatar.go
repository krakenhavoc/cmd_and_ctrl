package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// East Wind Avatar — {3}{W} Creature — Bird Spirit Avatar 2/4 (Teenage
// Mutant Ninja Turtles, #4457):
//
//	"Flying, vigilance
//	 Alliance — Whenever another creature you control enters, this
//	 creature gets +1/+0 until end of turn."
//
// Alliance is an ability word (CR 207.2c), so the trigger is
// WheneverAnotherCreatureEntersUnderYourControl; the body is the shared
// thisGetsUntilEndOfTurn on the trigger's own source.
//
// No simplification.
func init() {
	const label = "East Wind Avatar — alliance: +1/+0 until end of turn"
	Register(Spec{
		OracleID:        "e2ef355a-2472-47e0-9b2e-987994291921",
		Name:            "East Wind Avatar",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Triggered: []game.TriggeredAbility{
			WheneverAnotherCreatureEntersUnderYourControl(label, thisGetsUntilEndOfTurn(1, 0, label)),
		},
	})
}
