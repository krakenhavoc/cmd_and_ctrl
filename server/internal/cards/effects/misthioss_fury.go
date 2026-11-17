package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Misthios's Fury — Instant {1}{R} (Assassin's Creed, #5466):
//
//	"Misthios's Fury deals 3 damage to target creature. If you control
//	 an Equipment, Misthios's Fury also deals 2 damage to that
//	 creature's controller."
//
// Judgment Bolt's shape with a fixed 2: the creature's controller is
// read as the spell resolves, before the damage, and the Equipment is
// counted then too. An illegal target means the spell does nothing
// (CR 608.2b), so the controller takes nothing either.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e99c088e-1326-4844-b932-fb51802656f5",
		Name:         "Misthios's Fury",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(DamageToTarget(0, 3)),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			target, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			victim, ok := ctx.Game.LookupCardForEffect(target)
			if !ok {
				return nil
			}
			if err := (DealDamage{Source: ctx.Source(), Target: target, Amount: 3}).Apply(ctx); err != nil {
				return err
			}
			if countControlled(ctx.Game, item.Controller, func(c game.Card) bool { return c.HasSubtype("Equipment") }) == 0 {
				return nil
			}
			return DealDamage{Source: ctx.Source(), Target: victim.Controller, Amount: 2}.Apply(ctx)
		},
	})
}
