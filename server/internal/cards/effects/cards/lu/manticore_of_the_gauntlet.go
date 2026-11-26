package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Manticore of the Gauntlet — Creature — Manticore {4}{R}, 5/4 (Amonkhet, S82 OLD batch 242, #5535):
//
//	"When this creature enters, put a -1/-1 counter on target creature you
//	 control. This creature deals 3 damage to target opponent or planeswalker."
//
// One enters trigger with two targets (CR 115.1): a creature you control, which
// may be the Manticore, and an opponent or planeswalker. Each is judged on its
// own as the ability resolves (CR 608.2b); the damage is dealt by the Manticore.
//
// No simplification.
func init() {
	t := effects.WhenThisEnters("Manticore of the Gauntlet — a -1/-1 counter on target creature you control, 3 damage to target opponent or planeswalker",
		func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetCard {
				if err := (effects.AddCounter{Target: t.ID, Kind: game.CounterMinusOne, N: 1}).Apply(ctx); err != nil {
					return err
				}
			}
			victim, ok := ctx.ClauseTarget(1)
			if !ok {
				return nil
			}
			return effects.DealDamage{Source: ctx.Source(), Target: victim.ID, Amount: 3}.Apply(ctx)
		})
	t.Targets = effects.Clauses(effects.TargetCreatureYouControl(), effects.TargetOpponentOrPlaneswalker())
	effects.Register(effects.Spec{
		OracleID:     "3ec1b67a-5ff8-4ce7-8388-d11a23f76f0e",
		Name:         "Manticore of the Gauntlet",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{t},
	})
}
