package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lifesmith — Creature — Human Artificer {1}{G}, 2/1 (Scars of Mirrodin, S82
// OLD batch 472, #5577):
//
//	"Whenever you cast an artifact spell, you may pay {1}. If you do, you gain
//	 3 life."
//
// Myrsmith's cast trigger (OldB472WheneverYouCastAnArtifactMayPay): the {1} is
// asked as it resolves and the life is gained only once it has been paid.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "95f5c0ac-8835-4b31-863b-84f5273aac50",
		Name:         "Lifesmith",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OldB472WheneverYouCastAnArtifactMayPay("Lifesmith", "you may pay {1} to gain 3 life", "pay {1} to gain 3 life?", lifesmithPaid),
		},
	})
}

var lifesmithPaid = effects.PayUnlessBranches("pay-unless/s82-old-b472-lifesmith-life",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.GainLife{Player: ctx.Controller(), Amount: 3}.Apply(ctx)
	}, nil)
