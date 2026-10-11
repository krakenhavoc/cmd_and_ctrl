package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Zulaport Cutthroat — 1/1 Creature — Human Rogue Ally for {1}{B}:
//
//	"Whenever this creature or another creature you control dies,
//	each opponent loses 1 life and you gain 1 life."
//
// Blood Artist's untargeted cousin, narrower in what triggers it
// (your creatures only) and wider in what it hits (every opponent).
// You gain exactly 1 regardless of the table size — the life gain
// isn't per-opponent.
func init() {
	Register(Spec{
		OracleID:     "76b003e0-15af-4f22-bdf2-1ade5430964a",
		Name:         "Zulaport Cutthroat",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(WheneverACreatureYouControlDies("Zulaport Cutthroat — each opponent loses 1", Do(EachOpponentLosesLife{N: 1}, GainLife{Amount: 1})), game.Purpose{DeathPayoff: true}),
		},
	})
}
