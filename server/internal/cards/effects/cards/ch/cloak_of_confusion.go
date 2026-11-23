package cardsch

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:     "5d258409-28d3-4eac-9f1a-3e494db818bf",
		Name:         "Cloak of Confusion",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(effects.YouControl()),
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventBlockersDeclared},
				AppliesTo: effects.EnchantedCreatureAttacksAndIsNotBlocked,
				Key:       "Cloak of Confusion — you may have it assign no combat damage; if you do, defending player discards a card at random",
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					host, defender := effects.EnchantedCreatureOf(ctx), effects.TriggeringActor(item)
					if host == uuid.Nil {
						return nil
					}
					return effects.MayChoice{
						Question: "Cloak of Confusion — have the enchanted creature assign no combat damage this turn? (if you do, defending player discards a card at random)",
						Then:     cloakOfConfusionAnswer,
						Carry:    []uuid.UUID{host, defender},
					}.Apply(ctx)
				},
			},
		},
	})
}

// cloakOfConfusionAnswer is the "yes": the enchanted creature (the first
// carried ID) assigns no combat damage, and the defending player (the
// second) discards a card at random. Keyed (#4291).
var cloakOfConfusionAnswer = effects.MayYes("confirm/cloak-of-confusion-no-damage", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	host, defender := effects.CarryAt(r, 0), effects.CarryAt(r, 1)
	effects.EnchantedAssignsNoCombatDamage(ctx, host, "Cloak of Confusion — assigns no combat damage this turn")
	if !effects.DefendingPlayerStillIn(ctx.Game, defender) {
		return nil
	}
	return effects.DiscardCards{Player: defender, N: 1}.Apply(ctx)
})
