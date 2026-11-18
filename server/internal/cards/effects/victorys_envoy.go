package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Victory's Envoy — Creature — Human Cleric {3}{W}{W}, 3/3 (Commander
// Legends: Battle for Baldur's Gate reprint):
//
//	"At the beginning of your upkeep, put a +1/+1 counter on each other
//	 creature you control."
//
// The set of creatures is read once as the trigger resolves; the Envoy
// gets none (Bellowing Aegisaur's body).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fbdff913-8f86-4fcb-87c3-de2011fd036d",
		Name:         "Victory's Envoy",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Victory's Envoy — a +1/+1 counter on each other creature you control", putCounterOnEachOtherCreatureYouControl),
		},
	})
}
