package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Insatiable Frugivore — Creature — Rat Berserker {3}{B}, 2/4 (Bloomburrow
// Commander, #5368):
//
//	"When this creature enters, create a Food token, then you may exile
//	 three cards from your graveyard. If you do, repeat this process.
//	 {3}{B}, Sacrifice X Foods: Creatures you control get +X/+0 and gain
//	 menace until end of turn."
//
// The enters trigger makes a Food, then asks once whether to exile three
// cards from the graveyard (offered only with three to exile); a yes runs
// the same round again, so the loop ends the first time the controller
// declines or cannot pay. The activated ability announces X as the number of
// Foods sacrificed (SacrificeX); the set of creatures is fixed on
// resolution (CR 611.2c). X of 0 is not offered.
func init() {
	const label = "{3}{B}, Sacrifice X Foods: Creatures you control get +X/+0 and gain menace until end of turn."
	Register(Spec{
		OracleID:     "4f31eea0-61b9-467d-8393-0ab9f1b6e882",
		Name:         "Insatiable Frugivore",
		Completeness: CompletenessFull,
		XMatters:     true,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Insatiable Frugivore — create a Food token, then you may exile three cards from your graveyard; if you do, repeat this process",
				func(g *game.Game, item *game.StackItem) error {
					return frugivoreRound(NewContext(g, item))
				}),
		},
		Activated: []ActivatedAbility{{
			Label:   label,
			Cost:    Plus(ManaCost("{3}{B}"), SacrificeX("X Foods", cmdr2S18IsFood)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				x := ctx.X()
				if x <= 0 {
					return nil
				}
				mine := And(Creature(), YouControl())
				if err := (BoostUntilEOT{Match: mine, Power: x, Label: "Insatiable Frugivore — +X/+0"}).Apply(ctx); err != nil {
					return err
				}
				return GrantKeywordUntilEOT{Match: mine, Keywords: []string{"menace"}, Label: "Insatiable Frugivore — menace"}.Apply(ctx)
			},
		}},
	})
}

// frugivoreRound is one pass of the enters process: a Food, then the offer.
func frugivoreRound(ctx *Context) error {
	if err := (CreateToken{Template: FoodToken(), N: 1}).Apply(ctx); err != nil {
		return err
	}
	return MayPay{
		Chooser:  ctx.Controller(),
		Pay:      ExileFromGraveyardPayment(3),
		Question: "Insatiable Frugivore — exile three cards from your graveyard to repeat this process?",
		Then:     frugivoreRepeat,
	}.Apply(ctx)
}

// frugivoreRepeat is the paid branch: the process again. Assigned in init
// because the body reaches frugivoreRound, which reads this variable.
var frugivoreRepeat game.PayUnlessElse

func init() {
	frugivoreRepeat = PayUnlessBranches("pay-unless/cmdr2-s18-insatiable-frugivore-repeat",
		func(ctx *Context, _ game.PayUnlessAnswer) error { return frugivoreRound(ctx) }, nil)
}
