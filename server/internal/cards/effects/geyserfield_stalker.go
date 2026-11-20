package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Geyserfield Stalker — Creature — Elemental {4}{B}, 3/2 (Battle for
// Zendikar, #5547):
//
//	"Menace (This creature can't be blocked except by two or more
//	 creatures.)
//	 Landfall — Whenever a land you control enters, this creature
//	 gets +2/+2 until end of turn."
//
// Canopy Baloth's landfall on a creature with menace. Landfall is the
// stack trigger on a land entering under its controller's control (CR
// 207.2c), so a land an opponent plays does nothing; the pump ends in
// the cleanup step (CR 514.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d93d98ef-a5b6-4d6f-9828-fd5943471d0f",
		Name:            "Geyserfield Stalker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			Landfall("Geyserfield Stalker — +2/+2 until end of turn (landfall)",
				thisGetsUntilEndOfTurn(2, 2, "Geyserfield Stalker — +2/+2 until end of turn")),
		},
	})
}
