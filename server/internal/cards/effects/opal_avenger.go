package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Opal Avenger — Enchantment {2}{W}:
//
//	"When you have 10 or less life, if this permanent is an enchantment,
//	 it becomes a 3/5 Soldier creature."
//
// ADR 0107 PR 1 (#1858). A CR 603.8 state trigger with an intervening
// "if" (CR 603.4): it triggers as soon as its controller is at 10 life or
// less while it is an enchantment, and it checks that it is still an
// enchantment as it resolves. The creature it becomes is not an
// enchantment (CR 205.1a), so it does not trigger again. It stays white,
// and it stays a creature for as long as it is this object.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "13a85c9f-f653-4d90-bea6-94022ee82527",
		Name:         "Opal Avenger",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			whenStateIfThisIsAnEnchantment("Opal Avenger — it becomes a 3/5 Soldier creature",
				func(g *game.Game, _ *game.Card, controller uuid.UUID) bool {
					p := g.PlayerByIDForEffect(controller)
					return p != nil && !p.Eliminated && p.Life <= 10
				},
				thisEnchantmentBecomesACreature("Soldier", 3, 5, "Opal Avenger — a 3/5 Soldier creature")),
		},
	})
}
