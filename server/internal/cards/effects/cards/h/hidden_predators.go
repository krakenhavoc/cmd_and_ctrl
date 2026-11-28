package cardsh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hidden Predators — Enchantment {G}:
//
//	"When an opponent controls a creature with power 4 or greater, if this
//	 permanent is an enchantment, it becomes a 4/4 Beast creature."
//
// A CR 603.8 state trigger with an intervening "if" (CR 603.4): it triggers as
// soon as an opponent controls such a creature while this is an enchantment,
// and checks again as it resolves. The creature it becomes is not an
// enchantment, so it triggers once.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3b143227-f373-4403-b246-ee1c9a0dc85a",
		Name:         "Hidden Predators",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenStateIfThisIsAnEnchantment("Hidden Predators — it becomes a 4/4 Beast creature",
				func(g *game.Game, _ *game.Card, controller uuid.UUID) bool {
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.IsCreature() && c.Controller != controller && c.CurrentPower() >= 4 {
							return true
						}
					}
					return false
				},
				effects.ThisEnchantmentBecomesACreature("Beast", 4, 4, "Hidden Predators — a 4/4 Beast creature")),
		},
	})
}
