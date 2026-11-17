package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Elenda and Azor — Legendary Creature — Vampire Knight Sphinx
// {3}{W}{U}{B}, 6/6:
//
//	"Flying, ward {2}
//	 Whenever Elenda and Azor attacks, you may pay {X}{W}{U}{B}. If you do,
//	 draw X cards.
//	 At the beginning of each end step, you may pay 4 life. If you do,
//	 create a number of 1/1 black Vampire Knight creature tokens with
//	 lifelink equal to the number of cards you've drawn this turn."
//
// The attack ability is Stark's Ingenuity's pay-{X} shape with the three
// colored symbols printed beside the X (CR 107.3: X is chosen as the ability
// resolves and then the whole cost is offered).
//
// The end-step ability fires at every player's end step (CR 513.1), and its
// life payment is an ordinary "you may pay" that a player who can't pay
// (CR 119.4) is never offered. The number of tokens is read as the payment
// is made, from the turn's draw tally, so a draw in response counts and a
// turn with no draws pays 4 life for nothing, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4b9f463f-3786-4374-bf77-739ae162b3ba",
		Name:            "Elenda and Azor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{2}"), "Elenda and Azor — ward {2}"),
			WheneverThisAttacks("Elenda and Azor — you may pay {X}{W}{U}{B}; if you do, draw X cards", elendaAndAzorAttack),
			On(game.EventBeginEndStep, AnyPlayer, "Elenda and Azor — you may pay 4 life; if you do, create a number of 1/1 black Vampire Knight creature tokens with lifelink equal to the number of cards you've drawn this turn", elendaAndAzorEndStep),
		},
	})
}

func elendaAndAzorAttack(g *game.Game, item *game.StackItem) error {
	return MayPayX{
		Cost:  "{X}{W}{U}{B}",
		Label: "Elenda and Azor",
		Buys:  "draw X cards",
		Goal:  XAsHighAsYouCan,
		Unit:  game.PayAmountCards,
		Then:  elendaAndAzorAttackPaid,
	}.Apply(NewContext(g, item))
}

var elendaAndAzorAttackPaid = PayUnlessBranches("pay-unless/s82-mom-b21-elenda-and-azor-draw-x",
	func(ctx *Context, r game.PayUnlessAnswer) error {
		return DrawCards{N: r.Number}.Apply(ctx)
	}, nil)

func elendaAndAzorEndStep(g *game.Game, item *game.StackItem) error {
	return MayPay{
		Chooser:  item.Controller,
		Pay:      LifePayment(4),
		Question: "Elenda and Azor — pay 4 life to create a 1/1 black Vampire Knight with lifelink for each card you've drawn this turn?",
		Then:     elendaAndAzorEndPaid,
	}.Apply(NewContext(g, item))
}

var elendaAndAzorEndPaid = PayUnlessBranches("pay-unless/s82-mom-b21-elenda-and-azor-vampire-knights",
	func(ctx *Context, r game.PayUnlessAnswer) error {
		n := ctx.Game.TurnTallyFor(ctx.Controller()).CardsDrawn
		if n <= 0 {
			return nil
		}
		return CreateToken{Template: TokenCard("1/1 black Vampire Knight with lifelink"), N: n}.Apply(ctx)
	}, nil)
