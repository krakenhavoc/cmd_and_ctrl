package cardso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Oakhame Adversary — Creature — Elf Warrior {3}{G}, 2/3 (Throne of
// Eldraine, #5516):
//
//	"This spell costs {2} less to cast if an opponent controls a green
//	 permanent.
//	 Deathtouch
//	 Whenever this creature deals combat damage to a player, draw a card."
//
// The discount is judged for the caster as the cost is determined (CR
// 601.2f), reading the colours of each permanent an opponent controls
// through the layers.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3afe4852-d506-45d7-8a19-ddcbc3dd90d4",
		Name:            "Oakhame Adversary",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLess(2, "This spell costs {2} less to cast if an opponent controls a green permanent.",
				func(q game.CostQuery) bool {
					if q.Game == nil {
						return false
					}
					for _, c := range q.Game.BattlefieldCardsForEffect() {
						if c.Controller != q.Controller && c.Controller != uuid.Nil && c.HasColorWord("G") {
							return true
						}
					}
					return false
				}),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, effects.ThisDealtCombatDamageToAPlayer,
				"Oakhame Adversary — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
