package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Foul-Tongue Invocation — Instant {2}{B} (Dragons of Tarkir, #5551):
//
//	"As an additional cost to cast this spell, you may reveal a Dragon card from your hand.
//	 Target player sacrifices a creature of their choice. If you revealed a Dragon card or controlled a Dragon as you cast this spell, you gain 4 life."
//
// Draconic Roar's optional reveal and cast-time Dragon check, with Chainer's
// Edict's sacrifice. The life is gained after the sacrifice, whether or not
// the player had a creature to give.
//
// No simplification.

func init() {
	const dragonKey = "controlled_a_dragon"
	effects.Register(effects.Spec{
		OracleID:      "fbf27d1b-2b2a-49f3-abcd-20a6c79c44ee",
		Name:          "Foul-Tongue Invocation",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.OptionalBehold("a", "Dragon")},
		CastTimeValues: []game.CastTimeValue{{
			Key:   dragonKey,
			Value: effects.ControlsADragonAsCast,
		}},
		Targets: effects.TargetPlayer("target player"),
		Purpose: game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.TargetPlayerSacrificesACreature("Foul-Tongue Invocation — sacrifice a creature")(ctx.Game, item); err != nil {
				return err
			}
			controlled, _ := ctx.CastValue(dragonKey)
			if !ctx.WasBeheld() && controlled == 0 {
				return nil
			}
			return effects.GainLife{Amount: 4}.Apply(ctx)
		},
	})
}
