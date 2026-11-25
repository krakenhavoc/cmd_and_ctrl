package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Snapdax, Apex of the Hunt — Legendary Creature — Dinosaur Cat Nightmare
// {1}{R}{W}{B}, 3/5 (Ikoria: Lair of Behemoths, #5512):
//
//	"Mutate {2}{B/R}{W}{W} (If you cast this spell for its mutate cost, put it
//	 over or under target non-Human creature you own. They mutate into the
//	 creature on top plus all abilities from under it.)
//	 Double strike
//	 Whenever this creature mutates, it deals 4 damage to target creature or
//	 planeswalker an opponent controls and you gain 4 life."
//
// Mutate is the engine's (ADR 0223). The damage is dealt by the merged
// permanent, and the life is gained in the same resolution; with the target
// illegal as it resolves neither happens (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "b2b78623-5a10-4ef7-b983-3d82478269c3",
		Name:             "Snapdax, Apex of the Hunt",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"double strike"},
		AlternativeCosts: []game.AlternativeCost{effects.Mutate("{2}{B/R}{W}{W}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventMutate, effects.ThisCreatureMutates,
				"Snapdax, Apex of the Hunt — it deals 4 damage to target creature or planeswalker an opponent controls and you gain 4 life",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					if err := (effects.DealDamage{Source: ctx.Source(), Target: id, Amount: 4}).Apply(ctx); err != nil {
						return err
					}
					return effects.GainLife{Amount: 4}.Apply(ctx)
				}),
				effects.Harmful(effects.TargetPermanent("target creature or planeswalker an opponent controls", effects.And(effects.Or(effects.Creature(), effects.Planeswalker()), effects.OpponentControls())))),
		},
	})
}
