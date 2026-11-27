package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Forcemage Advocate — Creature — Centaur Shaman {1}{G}, 2/1 (Judgment):
//
//	"{T}: Return target card from an opponent's graveyard to their hand. Put
//	 a +1/+1 counter on target creature."
//
// Shieldmage Advocate's two target clauses (ADR 0108 §7): each is read as the
// ability resolves, and one that has become illegal is skipped while the other
// still happens (CR 608.2b). The card goes to its owner's hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c8c5cd28-4f27-424d-92bc-e7d6f15a5f9f",
		Name:         "Forcemage Advocate",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label: "{T}: Return target card from an opponent's graveyard to their hand. Put a +1/+1 counter on target creature.",
			Cost:  effects.TapCost(),
			Targets: effects.Clauses(
				effects.TargetCardInGraveyard("target card from an opponent's graveyard", effects.NotOwnedByCaster()),
				effects.TargetCreature("target creature"),
			),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetCard && ctx.IsTargetLegal(t) {
					if err := (effects.ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneHand}).Apply(ctx); err != nil {
						return err
					}
				}
				t, ok := ctx.ClauseTarget(1)
				if !ok || t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
					return nil
				}
				return effects.AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
			},
		}},
	})
}
