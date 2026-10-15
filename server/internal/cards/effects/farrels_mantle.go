package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Farrel's Mantle — Enchantment — Aura {2}{W}:
//
//	"Enchant creature
//	 Whenever enchanted creature attacks and isn't blocked, its controller
//	 may have it deal damage equal to its power plus 2 to another target
//	 creature. If that player does, the attacking creature assigns no combat
//	 damage this turn."
//
// #3531: the trigger is the Aura's own and so is its target (another creature
// than the enchanted one, chosen by the Aura's controller), but the "may" is
// the enchanted creature's controller's, who may be someone else. The damage
// is dealt by the enchanted creature, with its power as the ability resolves.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "ffae3ff4-67f7-4e41-b166-c63f88c7aa08",
		Name:         "Farrel's Mantle",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Triggered: []game.TriggeredAbility{
			{
				Watches:     []game.EventKind{game.EventBlockersDeclared},
				AppliesTo:   enchantedCreatureAttacksAndIsNotBlocked,
				TargetsFrom: farrelsMantleTarget,
				Key:         "Farrel's Mantle — its controller may have it deal damage equal to its power plus 2 to another target creature",
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					host := enchantedCreatureOf(ctx)
					hc, ok := g.LookupCardForEffect(host)
					if host == uuid.Nil || !ok {
						return nil
					}
					target := firstLegalCardTarget(ctx)
					return MayChoice{
						Player:   hc.Controller,
						Question: "Farrel's Mantle — have the enchanted creature deal damage equal to its power plus 2 to the target creature? (if you do, it assigns no combat damage this turn)",
						OnYes: func(ctx *Context) error {
							amount := b31CurrentPowerOf(ctx.Game, host) + 2
							if err := (DealDamage{Source: host, Target: target, Amount: amount}).Apply(ctx); err != nil {
								return err
							}
							enchantedAssignsNoCombatDamage(ctx, host, "Farrel's Mantle — assigns no combat damage this turn")
							return nil
						},
					}.Apply(ctx)
				},
			},
		},
	})
}

// farrelsMantleTarget is "another target creature": any creature but the
// one the Aura enchants.
func farrelsMantleTarget(_ game.TriggerContext, source *game.Card, _ *game.Game) *game.TargetSpec {
	host := source.AttachedTo.ID
	return &game.TargetSpec{
		Mode: "creature", Label: "another target creature",
		Zones: []game.ZoneKind{game.ZoneBattlefield},
		CardOK: func(_ *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool {
			return c.IsCreature() && c.InstanceID != host
		},
		Min: 1, Max: 1,
	}
}
