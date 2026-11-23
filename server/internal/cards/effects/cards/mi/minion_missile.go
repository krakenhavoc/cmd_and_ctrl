package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Minion Missile — Sorcery {1}{B}:
//
//	"As an additional cost to cast this spell, sacrifice a creature or
//	 discard a card. Destroy target creature. Minion Missile deals 2
//	 damage to that creature's controller."
//
// The either/or cost (ADR 0100 §2). "That creature's controller" is read
// before the destroy, while the creature is still on the battlefield —
// its controller as it last existed there (CR 608.2h) — so an
// indestructible or regenerated creature's controller is hit too: the
// damage does not depend on the destroy succeeding. A target that has
// become illegal counters the spell (CR 608.2b), and nothing is dealt.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "670cf656-5c72-4d28-b5e6-dc77781dc001",
		Name:         "Minion Missile",
		Completeness: effects.CompletenessFull,
		AdditionalCost: effects.EitherCost(
			effects.SacrificeCost("a creature", effects.Creature()).Keyed("sacrifice"),
			effects.DiscardCost(1).Keyed("discard"),
		),
		Targets: effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			for _, ref := range ctx.LegalTargets() {
				if ref.Kind != game.TargetCard {
					continue
				}
				victim, ok := ctx.Game.LookupCardForEffect(ref.ID)
				if !ok {
					return nil
				}
				if err := (effects.DestroyTarget{Target: ref.ID}).Apply(ctx); err != nil {
					return err
				}
				return effects.DealDamage{Source: ctx.Source(), Target: victim.Controller, Amount: 2}.Apply(ctx)
			}
			return nil
		},
	})
}
