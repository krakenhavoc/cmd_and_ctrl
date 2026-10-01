package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lurking Jackals — Enchantment {B}:
//
//	"When an opponent has 10 or less life, if this permanent is an
//	 enchantment, it becomes a 3/2 Jackal creature."
//
// ADR 0107 PR 1 (#1858). A CR 603.8 state trigger with an intervening
// "if" (CR 603.4), like Opal Avenger's: any one opponent still in the
// game at 10 life or less is enough. The creature it becomes is not an
// enchantment (CR 205.1a), so it does not trigger again; it stays black.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e9e5e17a-95fc-41b7-af4f-aec4802960a0",
		Name:         "Lurking Jackals",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			whenStateIfThisIsAnEnchantment("Lurking Jackals — it becomes a 3/2 Jackal creature",
				func(g *game.Game, _ *game.Card, controller uuid.UUID) bool {
					for _, p := range livePlayers(g) {
						if p.ID != controller && p.Life <= 10 {
							return true
						}
					}
					return false
				},
				thisEnchantmentBecomesACreature("Jackal", 3, 2, "Lurking Jackals — a 3/2 Jackal creature")),
		},
	})
}
