package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Viashino Heretic — Creature — Lizard {2}{R} 1/3:
//
//	"{1}{R}, {T}: Destroy target artifact. This creature deals damage to
//	 that artifact's controller equal to the artifact's mana value."
//
// The artifact's controller and mana value are read before it is destroyed
// (CR 608.2h), and the damage is dealt whether or not it was destroyed: an
// indestructible artifact, or one regenerated, still hurts its controller. If
// the target is illegal on resolution the whole ability is countered (CR
// 608.2b) and nothing is dealt.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "36896550-151d-472c-bfc1-1d9bbf7cc04d",
		Name:         "Viashino Heretic",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}, {T}: Destroy target artifact. This creature deals damage to that artifact's controller equal to the artifact's mana value.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.Plus(effects.ManaCost("{1}{R}"), effects.TapCost()),
			Targets: effects.TargetPermanent("target artifact", effects.Artifact()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				legal := ctx.LegalTargets()
				if len(legal) == 0 {
					return nil
				}
				art, ok := g.LookupCardForEffect(legal[0].ID)
				if !ok {
					return nil
				}
				controller, mv := art.Controller, art.ManaValue()
				destroy := effects.DestroyTarget{Target: art.InstanceID}
				// The damage is a value read before the destruction, not a
				// gate on it (ADR 0013 section 5m item 5).
				if mv <= 0 {
					return destroy.Apply(ctx)
				}
				if err := destroy.Apply(ctx); err != nil {
					return err
				}
				return effects.DealDamage{Source: item.SourceCardID, Target: controller, Amount: mv}.Apply(ctx)
			},
		}},
	})
}
