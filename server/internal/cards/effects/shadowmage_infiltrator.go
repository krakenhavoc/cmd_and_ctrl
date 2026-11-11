package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shadowmage Infiltrator — Creature {1}{U}{B}, 1/3, Fear:
//
//	"Fear
//	 Whenever this creature deals combat damage to a player, you may draw a card."
func init() {
	Register(Spec{
		OracleID:        "3a6f886b-2043-47e9-9c0f-f7913a6fa67d",
		Name:            "Shadowmage Infiltrator",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"fear"},
		Triggered: []game.TriggeredAbility{
			OptionalDraw(WheneverThisDealsCombatDamageToAPlayer("Shadowmage Infiltrator — you may draw a card",
				Do(DrawCards{N: 1})), "Draw a card?", 1),
		},
	})
}
