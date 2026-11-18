package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rending Flame — Instant {2}{R} (Innistrad: Crimson Vow, #5495):
//
//	"Rending Flame deals 5 damage to target creature or planeswalker. If
//	 that permanent is a Spirit, Rending Flame also deals 2 damage to that
//	 permanent's controller."
//
// A target that has become illegal counters the whole spell (CR 608.2b).
// "Is a Spirit" and "that permanent's controller" are read off the
// permanent before the damage, as it is when the spell resolves (it may
// die to the damage, and CR 608.2h reads what it last was), the type word
// as a word of this card's text (CR 612.2) and through the layers. The 2
// damage is dealt by Rending Flame too.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "24ac7e7e-b60d-4965-91c1-7ea7e6387737",
		Name:         "Rending Flame",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(DamageToTarget(0, 5)),
		Targets:      TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			var spiritOf uuid.UUID
			isSpirit := false
			if c, found := ctx.Game.LookupCardForEffect(t.ID); found {
				isSpirit = c.HasSubtypeWord(ctx.TextSubtype("Spirit"))
				spiritOf = c.Controller
			}
			if err := (DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 5}).Apply(ctx); err != nil {
				return err
			}
			if !isSpirit {
				return nil
			}
			return DealDamage{Source: ctx.Source(), Target: spiritOf, Amount: 2}.Apply(ctx)
		},
	})
}
