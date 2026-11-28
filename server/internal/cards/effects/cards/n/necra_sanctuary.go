package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Necra Sanctuary — Enchantment {2}{B} (Apocalypse, #5614):
//
//	"At the beginning of your upkeep, if you control a green or white
//	 permanent, target player loses 1 life. If you control a green permanent
//	 and a white permanent, that player loses 3 life instead."
//
// One triggered ability with an intervening if (CR 603.4), checked as it
// triggers and again on resolution; the target player is chosen as it goes
// on the stack and the amount is read as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2586a59d-8501-4c22-9d69-f4bf91de7024",
		Name:         "Necra Sanctuary",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.S82OldB634Sanctuary("G", "W", "Necra Sanctuary — target player loses 1 life (3 with a green and a white permanent)",
				effects.TargetPlayer("target player"), effects.S82OldB634TargetPlayerLosesLife),
		},
	})
}
