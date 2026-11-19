package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mystery Key — Artifact — Equipment {1}{U}:
//
//	"When equipped creature deals combat damage to a player, sacrifice
//	 this Equipment. If you do, draw three cards.
//	 Equip {1}"
//
// The combat-damage trigger is Bloodforged Battle-Axe's; the sacrifice and
// its "if you do" draw is SacrificeThisThen, so the three cards come only
// when the Key really left the battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6a1e0dfd-2dd1-4372-966e-e28b48ae0115",
		Name:         "Mystery Key",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attachedCreatureDealtCombatDamageToPlayer(ev, source, g)
			},
			Key: "Mystery Key — sacrifice this Equipment. If you do, draw three cards",
			Effect: SacrificeThisThen("then/mystery-key-draw-three", func(ctx *Context) error {
				return DrawCards{Player: ctx.Controller(), N: 3}.Apply(ctx)
			}),
		}},
		Activated: []ActivatedAbility{EquipAbility("{1}")},
	})
}
