package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Apothecary Initiate — Creature — Kithkin Cleric {W}, 1/1 (Shadowmoor, S82
// OLD b529, #5591):
//
//	"Whenever a player casts a white spell, you may pay {1}. If you do, you
//	 gain 1 life."
//
// Kor Firewalker's any-player cast trigger with a may-pay: the {1} is asked as
// the ability resolves (CR 603.5, 118.12) and the life comes only once it is
// paid. The Initiate's own cast does not trigger it: it is not on the
// battlefield yet.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a74d80ec-9f42-418d-ab98-897acc23c319",
		Name:         "Apothecary Initiate",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OldB529WhenAPlayerCasts("W", "Apothecary Initiate — you may pay {1}; if you do, you gain 1 life",
				effects.OldB529MayPayOne("Apothecary Initiate — pay {1} to gain 1 life?", apothecaryInitiatePaid)),
		},
	})
}

// apothecaryInitiatePaid is the paid branch. An on-disk key: never renamed,
// never reused.
var apothecaryInitiatePaid = effects.PayUnlessBranches("pay-unless/apothecary-initiate-gain-1",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.GainLife{Player: ctx.Controller(), Amount: 1}.Apply(ctx)
	}, nil)
