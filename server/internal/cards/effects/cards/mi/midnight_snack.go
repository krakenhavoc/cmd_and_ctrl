package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Midnight Snack — Enchantment {2}{B} (Foundations, #4465):
//
//	"Raid — At the beginning of your end step, if you attacked this turn,
//	 create a Food token.
//	 {2}{B}, Sacrifice this enchantment: Target opponent loses X life,
//	 where X is the amount of life you gained this turn."
//
// The Food is Searslicer Goblin's raid end-step shape. X is read as the
// ability resolves (CR 608.2h) off the turn's life-gained tally, which
// counts every gain, lifelink and Food included, and not a loss; the
// Enchantment is already sacrificed by then, and the tally does not care.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3ddcaee9-a11e-477a-bd19-f5ee10512549",
		Name:         "Midnight Snack",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginEndStep, effects.AllOf(effects.ByYou, effects.TmcB6YouAttackedThisTurn),
				"Midnight Snack — Raid: create a Food token",
				effects.Do(effects.If{Cond: effects.AvatarB11CondYouAttacked{}, Then: []effects.Applier{
					effects.CreateToken{Template: effects.FoodToken(), N: 1},
				}})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{B}, Sacrifice this enchantment: Target opponent loses X life, where X is the amount of life you gained this turn.",
			Cost:    effects.Plus(effects.ManaCost("{2}{B}"), effects.SacrificeThis()),
			Targets: effects.TargetPlayer("target opponent", effects.Opponent()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				victim, ok := effects.FirstLegalPlayerTarget(ctx)
				if !ok {
					return nil
				}
				return effects.GainLife{Player: victim, Amount: -effects.B15LifeGainedThisTurn(g, ctx.Controller())}.Apply(ctx)
			},
		}},
	})
}
