package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hunted Dragon — Creature — Dragon {3}{R}{R}, 6/6 (Ravnica: City of Guilds,
// #5600):
//
//	"Flying, haste
//	 When this creature enters, target opponent creates three 2/2 white
//	 Knight creature tokens with first strike."
//
// Hunted Horror's targeted enters trigger: the opponent creates the
// Knights, so they own and control them (CR 111.2). An opponent who has
// left by the time it resolves gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "39a4416f-9b40-432e-aeb6-cc7bf519b37a",
		Name:            "Hunted Dragon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Triggered: []game.TriggeredAbility{
			ravB4TargetOpponentCreates("Hunted Dragon — target opponent creates three 2/2 white Knights with first strike", "2/2 white Knight with first strike", 3),
		},
	})
}
