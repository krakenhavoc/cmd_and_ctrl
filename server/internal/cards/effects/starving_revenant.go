package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Starving Revenant — Creature — Spirit Horror {2}{B}{B}, 4/4:
//
//	"When this creature enters, surveil 2. Then for each card you put
//	 on top of your library, you draw a card and you lose 3 life.
//	 Descend 8 — Whenever you draw a card, if there are eight or more
//	 permanent cards in your graveyard, target opponent loses 1 life
//	 and you gain 1 life."
//
// # The surveil
//
// "For each card you put on top" is a count of the cards the surveil
// kept: a keyed surveil's tail is told it (TailOutcome.Amount, #4253
// slice 10b). The draw and the life loss ride the tail, as Scry's
// warning says they must.
//
// Each kept card is its own pair: draw one, then lose 3. A draw from an
// empty library is still a draw attempt (CR 704.5b).
//
// # Descend 8
//
// A trigger on each of your draws, with an intervening if (CR 603.4)
// that counts permanent cards (CR 110.4) in your graveyard: asked as
// the draw happens, and again as the trigger resolves. Not a keyword
// to the engine, only a condition read off the board.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2ca969eb-3d79-4d1f-8d9d-7b8204ad166a",
		Name:         "Starving Revenant",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Starving Revenant — surveil 2, then draw and lose 3 life for each card kept on top",
				starvingRevenantSurveil),
			{
				Watches: []game.EventKind{game.EventDrawCard},
				Key:     "Starving Revenant — target opponent loses 1 life and you gain 1 life",
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return ev.Actor == source.Controller && permanentCardsInGraveyard(g, source.Controller) >= 8
				},
				Targets: TargetPlayer("target opponent", Opponent()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					if permanentCardsInGraveyard(g, item.Controller) < 8 {
						return nil
					}
					return drainTargetOpponentOne(g, item)
				},
			},
		},
	})
}

// starvingRevenantSurveil is the enters trigger: surveil 2, then a
// draw and 3 life lost per card put back on top. The rest is a keyed tail
// (#4253 slice 10b), told how many cards went back on top, so the surveil
// prompt is a restore point.
func starvingRevenantSurveil(g *game.Game, item *game.StackItem) error {
	me := item.Controller
	if g.PlayerByIDForEffect(me) == nil {
		return nil
	}
	ctx := NewContext(g, item)
	g.SurveilKeyedForEffect(me, item.SourceCardID, 2, ctx.ItemTail(starvingRevenantKept))
	return nil
}

// starvingRevenantKept is the surveil's "then": for each card kept on top
// (the outcome's Amount), a draw and 3 life lost. An on-disk key: never
// renamed, never reused.
var starvingRevenantKept = CardTail("then/starving-revenant-kept", func(ctx *Context, o game.TailOutcome) error {
	if ctx.Game.PlayerByIDForEffect(ctx.Controller()) == nil || o.Amount <= 0 {
		return nil
	}
	ctx.RecordStep(starvingRevenantRecord, nil, o.Amount)
	return starvingRevenantCard.Run(ctx)
})

// starvingRevenantRecord is how many cards are still to be drawn.
const starvingRevenantRecord = "starving-revenant/kept"

// starvingRevenantCard is one card kept on top: draw a card and lose 3 life,
// and then the next one. The loss waits for a prompt the draw asks (a dredge
// offer) and the next draw for one the loss asks (CR 608.2c, #4155). It runs
// itself again, so it is registered in init.
var starvingRevenantCard ThenSteps

func init() {
	starvingRevenantCard = RegisterThenSteps("then-steps/starving-revenant-card",
		Instruction(func(ctx *Context) error { return ctx.Game.DrawNForEffect(ctx.Controller(), 1) }),
		Instruction(func(ctx *Context) error {
			return ctx.Game.ChangePlayerLifeForEffect(ctx.Item.SourceCardID, ctx.Controller(), -3)
		}),
		Instruction(func(ctx *Context) error {
			left := ctx.StepRecord(starvingRevenantRecord).N() - 1
			if left <= 0 {
				return nil
			}
			ctx.RecordStep(starvingRevenantRecord, nil, left)
			return starvingRevenantCard.Run(ctx)
		}),
	)
}
