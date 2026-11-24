package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Arrogant Poet — {1}{B} Creature — Human Warlock 2/1 (Strixhaven: School of
// Mages, #5502):
//
//	"Whenever this creature attacks, you may pay 2 life. If you do, it gains
//	 flying until end of turn."
//
// The life is paid as the ability resolves (CR 118.12), as Zhentarim
// Bandit's is: a controller who can't pay 2 life (CR 119.4) is not asked,
// and declining gives nothing. "It" is the creature that attacked (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0b2e9571-0c28-4a8f-9ba7-0147b40290a7",
		Name:         "Arrogant Poet",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Arrogant Poet — you may pay 2 life; if you do, it gains flying until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Pay:      effects.LifePayment(2),
						Question: "Arrogant Poet — pay 2 life to gain flying until end of turn?",
						Then:     arrogantPoetPaid,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// arrogantPoetPaid is "if you do, it gains flying until end of turn". An
// on-disk key: never renamed, never reused.
var arrogantPoetPaid = effects.PayUnlessBranches("pay-unless/arrogant-poet-flying",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		if !effects.SourceIsStillThisPermanent(ctx.Game, ctx.Item) {
			return nil
		}
		return effects.GrantKeywordUntilEOT{Target: ctx.Item.SourceCardID, Keywords: []string{"flying"},
			Label: "Arrogant Poet — flying"}.Apply(ctx)
	}, nil)
