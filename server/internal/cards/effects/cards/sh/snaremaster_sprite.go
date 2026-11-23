package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Snaremaster Sprite — Creature — Faerie Wizard {U}, 1/1 (Wilds of
// Eldraine, #4471):
//
//	"Flying
//	 When this creature enters, you may pay {2}. When you do, tap target
//	 creature an opponent controls and put a stun counter on it."
//
// Tolarian Kraken's "you may pay. When you do" shape: MayPay asks as the
// trigger resolves, and only a payment queues the reflexive trigger,
// which then picks its target (ADR 0129). Its body is Deluge Virtuoso's
// shared tap-and-stun (tapTargetAndStunIt).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b0632f1b-14e5-4534-b2c5-af9d557fdcc0",
		Name:            "Snaremaster Sprite",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Snaremaster Sprite — you may pay {2}", func(g *game.Game, item *game.StackItem) error {
				return effects.MayPay{
					Chooser:  item.Controller,
					Cost:     "{2}",
					Question: "Snaremaster Sprite — pay {2} to tap target creature an opponent controls and put a stun counter on it?",
					Then:     snaremasterSpritePaid,
				}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}

var snaremasterSpritePaid = effects.PayUnlessBranches("pay-unless/snaremaster-sprite-tap-and-stun",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.WhenYouDo("Snaremaster Sprite — tap target creature an opponent controls and put a stun counter on it",
			snaremasterSpriteStunBody).Apply(ctx)
	}, nil)

var snaremasterSpriteStunBody = game.ReflexiveBody("snaremaster-sprite/tap-and-stun",
	effects.SimpleBody(effects.TapTargetAndStunIt),
	effects.ConstTargets(func() *game.TargetSpec {
		return effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())
	}))
