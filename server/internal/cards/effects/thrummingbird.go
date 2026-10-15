package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thrummingbird — Creature — Phyrexian Bird Horror {1}{U}, 1/1 (EDHREC
// top 1,000, #3654):
//
//	"Flying
//	 Whenever this creature deals combat damage to a player,
//	 proliferate."
//
// Bloated Contaminator's trigger on a flyer. Proliferate asks its
// controller which permanents and players to add a counter to (CR 701.34).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "eac94269-4baa-4b8e-a0fd-d6b227d1cde3",
		Name:            "Thrummingbird",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Thrummingbird — proliferate", Do(Proliferate{})),
		},
	})
}
