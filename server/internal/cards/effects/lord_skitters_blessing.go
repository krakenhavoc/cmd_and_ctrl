package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lord Skitter's Blessing — Enchantment {1}{B}:
//
//	"When this enchantment enters, create a Wicked Role token attached
//	 to target creature you control. (Enchanted creature gets +1/+0.
//	 When this token is put into a graveyard, each opponent loses 1
//	 life.)
//	 At the beginning of your draw step, if you control an enchanted
//	 creature, you lose 1 life and you draw an additional card."
//
// An "enchanted creature" is any creature with an Aura attached, the
// Role included (the same reading Ellivere uses). The draw-step clause
// is an intervening if (CR 603.4): checked when the step begins and
// again on resolution, so a Role that falls off in response leaves the
// trigger with nothing to do.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "d5bd40fc-6138-4752-b3ff-b1e5fa269214",
		Name:         "Lord Skitter's Blessing",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Lord Skitter's Blessing — create a Wicked Role token attached to target creature you control",
					createRoleOnFirstTarget(RoleWicked)),
				TargetCreature("target creature you control", YouControl())),
			{
				Watches: []game.EventKind{game.EventBeginDrawStep},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return ev.Actor == source.Controller && controlsEnchantedCreature(g, source.Controller)
				},
				Key: "Lord Skitter's Blessing — lose 1 life and draw an additional card",
				Effect: func(g *game.Game, item *game.StackItem) error {
					if !controlsEnchantedCreature(g, item.Controller) {
						return nil
					}
					ctx := NewContext(g, item)
					if err := g.ChangePlayerLifeForEffect(ctx.Source(), item.Controller, -1); err != nil {
						return err
					}
					return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
				},
			},
		},
	})
}

// controlsEnchantedCreature is "you control an enchanted creature".
func controlsEnchantedCreature(g *game.Game, controller uuid.UUID) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.IsCreature() && isEnchantedByAura(g, c.InstanceID) {
			return true
		}
	}
	return false
}
