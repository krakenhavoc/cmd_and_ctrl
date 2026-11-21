package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Najal, the Storm Runner — Legendary Creature — Efreet Wizard {2}{U}{U}{R},
// 5/4 (Dominaria United):
//
//	"You may cast sorcery spells as though they had flash.
//	 Whenever Najal attacks, you may pay {2}. If you do, when you next cast an
//	 instant or sorcery spell this turn, copy it. You may choose new targets
//	 for the copy."
//
// The first ability is Teferi, Time Raveler's, as a static (Gandalf, Friend of
// the Shire's shape). The attack trigger is MayPay: {2} is paid as the trigger
// resolves, and "if you do" is its paid branch, which makes Doublecast's delayed
// trigger (CR 603.7b, WhenYouNextCast with the copy-the-spell body). The delayed
// trigger belongs to the attacking player and ends with the turn if no instant or
// sorcery is cast. Declining, or being unable to pay, makes nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2a7504b9-220d-412e-9381-b6a8a3750241",
		Name:         "Najal, the Storm Runner",
		Completeness: effects.CompletenessFull,
		CastTimings: []game.CastTimingRule{
			effects.CastKindAsThoughFlash(game.PermissionFilter{SorceryOnly: true}, "You may cast sorcery spells as though they had flash."),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, effects.Self, "Najal, the Storm Runner — you may pay {2}; if you do, when you next cast an instant or sorcery spell this turn, copy it",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{2}",
						Question: "Najal, the Storm Runner — pay {2} to copy the next instant or sorcery spell you cast this turn?",
						Then:     najalPaid,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

var najalPaid = effects.PayUnlessBranches("najal-the-storm-runner/paid", func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
	return effects.WhenYouNextCast(
		"Najal, the Storm Runner — copy that spell",
		game.CastFilter{Types: []string{"Instant", "Sorcery"}},
		effects.CopyTheSpellBody,
	).Apply(ctx)
}, nil)
