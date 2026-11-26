package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Redcap Melee — Instant {R} (Throne of Eldraine, #5516):
//
//	"Redcap Melee deals 4 damage to target creature or planeswalker. If a
//	 nonred permanent is dealt damage this way, you sacrifice a land."
//
// The rider is the damage's continuation (ADR 0108 §1 decision 3): it runs
// only for a permanent that was dealt more than 0 damage, and reads its
// colour as the damage settles, so damage that was prevented costs no land.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1b78aa71-b554-45dc-a737-52cd74c3e7f1",
		Name:         "Redcap Melee",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 4)),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			t, ok := effects.FirstLegalTarget(ctx)
			if !ok {
				return nil
			}
			return effects.DealDamageThen(ctx, t.ID, 4, effects.ItemRecipient(item, redcapMeleeSacrifice))
		},
	})
}

// redcapMeleeSacrifice is "if a nonred permanent is dealt damage this way,
// you sacrifice a land". An on-disk key.
var redcapMeleeSacrifice = effects.RecipientTail("recipient/redcap-melee-sacrifice-a-land", func(ctx *effects.Context, target uuid.UUID, dealt int, _ game.TailOutcome) error {
	if dealt <= 0 {
		return nil
	}
	if z := ctx.Game.FindCardZoneForEffect(target); z == nil || z.Kind != game.ZoneBattlefield {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(target)
	if !ok || effects.OfColor("R")(ctx.Game, target, c) {
		return nil
	}
	ctx.Game.PlayerSacrificesForEffect(ctx.Source(), ctx.Controller(), effects.SacrificeSpec("a land", effects.Land()), "Redcap Melee — sacrifice a land")
	return nil
})
