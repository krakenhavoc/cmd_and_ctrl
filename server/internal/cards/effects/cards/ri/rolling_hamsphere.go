package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rolling Hamsphere — Artifact — Vehicle {7}, 4/4 (Bloomburrow Commander):
//
//	"This Vehicle gets +1/+1 for each Hamster you control.
//	 Whenever this Vehicle attacks, create three 1/1 red Hamster creature
//	 tokens, then it deals X damage to any target, where X is the number of
//	 Hamsters you control.
//	 Crew 3"
//
// The pump is a layer 7c self-pump counting the Hamsters its controller
// controls on every recompute. The target of the attack trigger is chosen as it
// is put on the stack (CR 603.3d); X is counted as it resolves, after the three
// tokens have entered (CR 608.2h), and the Vehicle is the damage source.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a6c3aa17-8338-4daa-889e-d36dc715e546",
		Name:         "Rolling Hamsphere",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 3},
		Static: []game.StaticAbility{{
			Label:    "This Vehicle gets +1/+1 for each Hamster you control.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := effects.Cmdr2B21HamstersYouControl(g, source.Controller)
				c.Power += n
				c.Toughness += n
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WheneverThisAttacks("Rolling Hamsphere — create three 1/1 Hamsters, then it deals X damage to any target, where X is the number of Hamsters you control",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if err := (effects.CreateToken{Template: effects.TokenCard("1/1 red Hamster"), N: 3}).Apply(ctx); err != nil {
						return err
					}
					return effects.SourceDealsComputedDamageToEachLegalTarget(func(c *effects.Context) int {
						return effects.Cmdr2B21HamstersYouControl(c.Game, c.Controller())
					})(g, item)
				}), effects.TargetAny()),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 3",
			Cost:    effects.CrewCost(3),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Rolling Hamsphere"),
		}},
	})
}
