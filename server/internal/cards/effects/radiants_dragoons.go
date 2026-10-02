package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Radiant's Dragoons — Creature — Human Soldier, {3}{W}, 2/5:
//
//	"Echo {3}{W} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
//	 When this creature enters, you gain 5 life."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a0fe91a8-633c-4902-bc75-da0a5aa76992",
		Name:         "Radiant's Dragoons",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Echo("Radiant's Dragoons", "{3}{W}"),
			WhenThisEnters("Radiant's Dragoons — gain 5 life", Do(GainLife{Amount: 5})),
		},
	})
}
