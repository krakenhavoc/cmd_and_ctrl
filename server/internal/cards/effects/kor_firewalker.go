package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kor Firewalker — Creature — Kor Soldier {W}{W}, 2/2 (Worldwake, #5582):
//
//	"Protection from red
//	 Whenever a player casts a red spell, you may gain 1 life."
//
// Angel's Feather's trigger for red, on a creature with protection
// from red (game/protection.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "fb37ff1c-7461-4a16-b27f-6eef0db937fe",
		Name:            "Kor Firewalker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"protection from red"},
		Triggered: []game.TriggeredAbility{
			Optional(On(game.EventCast, tmcB8APlayerCast(OfColor("R")),
				"Kor Firewalker — you may gain 1 life", Do(GainLife{Amount: 1})),
				"Kor Firewalker — gain 1 life?"),
		},
	})
}
