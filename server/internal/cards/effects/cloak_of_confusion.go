package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cloak of Confusion — Enchantment — Aura {1}{B}:
//
//	"Enchant creature you control
//	 Whenever enchanted creature attacks and isn't blocked, you may have it
//	 assign no combat damage this turn. If you do, defending player discards
//	 a card at random."
//
// #3531: the trigger is the Aura's own (CR 603.2), on the completed block
// declaration of the player the enchanted creature attacks. On yes the
// enchanted creature is marked by the CR 510.1a gate and the defending
// player discards at random (CR 701.9b).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "5d258409-28d3-4eac-9f1a-3e494db818bf",
		Name:         "Cloak of Confusion",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(YouControl()),
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventBlockersDeclared},
				AppliesTo: enchantedCreatureAttacksAndIsNotBlocked,
				Key:       "Cloak of Confusion — you may have it assign no combat damage; if you do, defending player discards a card at random",
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					host, defender := enchantedCreatureOf(ctx), triggeringActor(item)
					if host == uuid.Nil {
						return nil
					}
					return MayChoice{
						Question: "Cloak of Confusion — have the enchanted creature assign no combat damage this turn? (if you do, defending player discards a card at random)",
						OnYes: func(ctx *Context) error {
							enchantedAssignsNoCombatDamage(ctx, host, "Cloak of Confusion — assigns no combat damage this turn")
							if !defendingPlayerStillIn(ctx.Game, defender) {
								return nil
							}
							return DiscardCards{Player: defender, N: 1}.Apply(ctx)
						},
					}.Apply(ctx)
				},
			},
		},
	})
}
