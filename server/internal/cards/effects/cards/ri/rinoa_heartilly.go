package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rinoa Heartilly — Legendary Creature — Human Rebel Warlock
// {3}{G}{W}, 4/4 (Final Fantasy, #4462):
//
//	"When Rinoa Heartilly enters, create Angelo, a legendary 1/1 green
//	 and white Dog creature token.
//	 Angelo Cannon — Whenever Rinoa Heartilly attacks, another target
//	 creature you control gets +1/+1 until end of turn for each creature
//	 you control."
//
// The creatures are counted as the attack trigger resolves (CR 608.2h),
// Angelo and the target included. If the target is gone then, nothing
// happens (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8896c74b-d466-4c25-94cf-82b5e2b1fdef",
		Name:         "Rinoa Heartilly",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Rinoa Heartilly — create Angelo, a legendary 1/1 green and white Dog",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 green and white Legendary Dog named Angelo"), N: 1})),
			effects.Targeting(
				effects.On(game.EventAttack, effects.ThisAttacked,
					"Rinoa Heartilly — Angelo Cannon: another target creature you control gets +1/+1 until end of turn for each creature you control",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						target, ok := effects.B16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						n := effects.CountControlled(g, item.Controller, func(c game.Card) bool { return c.IsCreature() })
						return effects.BoostUntilEOT{Target: target, Power: n, Toughness: n, Label: "Angelo Cannon — +1/+1 for each creature you control"}.Apply(ctx)
					}),
				effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl()))),
		},
	})
}
