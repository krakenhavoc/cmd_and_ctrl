package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Decree of Justice — Sorcery {X}{X}{2}{W}{W} (Modern Horizons 3):
//
//	"Create X 4/4 white Angel creature tokens with flying.
//	 Cycling {2}{W}
//	 When you cycle this card, you may pay {X}. If you do, create X 1/1
//	 white Soldier creature tokens."
//
// The spell is the double-X token maker (the announced X, CR 107.3b). Cycling
// is the shared CR 702.29a ability, and the cycle trigger is Decree of Pain's
// graveyard watcher (the card is in the graveyard by the time it fires, CR
// 702.29c). Its X is not announced: it is chosen as the trigger resolves
// (CR 608.2d, MayPayX), from 0 to what the controller can pay, and the
// Soldiers are made only once {X} is paid.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "67a98359-6f02-4556-ad47-c01b643686d6",
		Name:         "Decree of Justice",
		XMatters:     true,
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return CreateToken{Template: TokenCard("4/4 white Angel with flying"), N: ctx.X()}.Apply(ctx)
		},
		Activated: []ActivatedAbility{Cycling("{2}{W}")},
		Triggered: []game.TriggeredAbility{
			InGraveyard(On(game.EventCycle, Self,
				"Decree of Justice — cycled: you may pay {X}; if you do, create X 1/1 white Soldier tokens",
				func(g *game.Game, item *game.StackItem) error {
					return MayPayX{
						Cost:  "{X}",
						Label: "Decree of Justice",
						Buys:  "create X 1/1 white Soldier creature tokens",
						Goal:  XAsHighAsYouCan,
						Then:  decreeOfJusticePaid,
					}.Apply(NewContext(g, item))
				})),
		},
	})
}

// decreeOfJusticePaid is the paid half, keyed: the X paid rides the prompt.
// An on-disk key: never renamed, never reused.
var decreeOfJusticePaid = PayUnlessBranches("pay-unless/decree-of-justice-soldiers",
	func(ctx *Context, r game.PayUnlessAnswer) error {
		return CreateToken{Template: TokenCard("1/1 white Soldier"), N: r.Number}.Apply(ctx)
	}, nil)
