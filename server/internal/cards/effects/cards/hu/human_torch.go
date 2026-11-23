package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Human Torch — Legendary Creature — Human Hero {3}{R}, 3/2 (Marvel Super
// Heroes Commander, #4455):
//
//	"At the beginning of combat on your turn, if you've cast a noncreature
//	 spell this turn, Human Torch gains flying, double strike, and haste
//	 until end of turn.
//	 Whenever Human Torch attacks, you may pay {R}{G}{W}{U}. If you do,
//	 until end of turn, whenever he deals combat damage to an opponent, he
//	 deals that much damage to each other opponent."
//
// The first ability has an intervening if (CR 603.4): checked as combat
// begins and again as it resolves. The second asks for the payment as it
// resolves (MayPay in the step it was asked in); paid, it creates a repeating
// "this turn" watch bound to Human Torch (whenThatCreatureHitsAPlayer, opponent
// form), so a double strike hits twice and each hit is its own copy of the
// damage. The damage is dealt by Human Torch, not combat damage, and skips the
// opponent just hit.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d55cb52d-b35c-4602-a518-cd51f155a2d1",
		Name:         "Human Torch",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventStepBegan, effects.AllOf(effects.StepBegan(game.StepBeginCombat, true), effects.FranklinCastNoncreature),
				"Human Torch — gains flying, double strike, and haste until end of turn",
				humanTorchGainsKeywords),
			effects.WheneverThisAttacks("Human Torch — pay {R}{G}{W}{U} for damage to each other opponent", func(g *game.Game, item *game.StackItem) error {
				return effects.MayPay{
					Chooser:    item.Controller,
					Cost:       "{R}{G}{W}{U}",
					Question:   "Human Torch — pay {R}{G}{W}{U}? If you do, whenever he deals combat damage to an opponent this turn, he deals that much damage to each other opponent.",
					Then:       humanTorchPaid,
					InThisStep: true,
				}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}

func humanTorchGainsKeywords(g *game.Game, item *game.StackItem) error {
	if g.CastTallyFor(item.Controller).Noncreature == 0 {
		return nil
	}
	ctx := effects.NewContext(g, item)
	if ctx.IsNewSourceObject(item.SourceCardID) || !effects.OnBattlefield(g, item.SourceCardID) {
		return nil
	}
	return effects.GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"flying", "double strike", "haste"},
		Label: "Human Torch — flying, double strike, and haste"}.Apply(ctx)
}

var humanTorchPaid = effects.PayUnlessBranches("pay-unless/human-torch-share-damage",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.WhenThatCreatureHitsAPlayer(ctx, ctx.Source(), "Human Torch — he deals that much damage to each other opponent",
			true, effects.DurationUntilEndOfTurn(ctx), effects.ThatMuchToEachOtherOpponentBody)
	}, nil)
