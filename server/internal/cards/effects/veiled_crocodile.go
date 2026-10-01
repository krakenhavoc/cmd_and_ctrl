package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Veiled Crocodile — Enchantment {2}{U}:
//
//	"When a player has no cards in hand, if this permanent is an
//	 enchantment, it becomes a 4/4 Crocodile creature."
//
// ADR 0107 PR 1 (#1858). A CR 603.8 state trigger with an intervening
// "if" (CR 603.4). "A player" is any player still in the game, its
// controller included. Because the engine asks after every event, a hand
// that is empty only for a moment inside a resolution ("discard your
// hand, then draw that many cards") triggers it, exactly as CR 603.8's
// own example says. The creature it becomes is not an enchantment
// (CR 205.1a), so it does not trigger again; it stays blue.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bdc3ff35-f8d8-477d-b042-94a9f8ed7243",
		Name:         "Veiled Crocodile",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			whenStateIfThisIsAnEnchantment("Veiled Crocodile — it becomes a 4/4 Crocodile creature",
				func(g *game.Game, _ *game.Card, _ uuid.UUID) bool {
					for _, p := range livePlayers(g) {
						if p.Hand == nil || p.Hand.Size() == 0 {
							return true
						}
					}
					return false
				},
				thisEnchantmentBecomesACreature("Crocodile", 4, 4, "Veiled Crocodile — a 4/4 Crocodile creature")),
		},
	})
}
