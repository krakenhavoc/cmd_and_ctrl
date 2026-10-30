package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// theRuinousWreckingCrewOracle is the card's catalog key.
const theRuinousWreckingCrewOracle = "0ff9c0cc-4fb3-4117-9a8e-5b51f62306c4"

// The Ruinous Wrecking Crew — Legendary Creature — Human Villain
// {X}{B}{R}, 2/2 (Marvel Super Heroes, #4455):
//
//	"The Ruinous Wrecking Crew enters with X +1/+1 counters on it.
//	 When The Ruinous Wrecking Crew enters, choose up to X —
//	 • Discard a card, then draw a card.
//	 • Target opponent loses 2 life.
//	 • Destroy target token.
//	 • Each player sacrifices a creature of their choice."
//
// The counters are the X entry replacement (CR 614.1c, XCounters). The
// trigger's count is the X paid for the spell (CR 107.3m), read off the
// entering creature as the trigger goes on the stack (ChooseUpToX,
// #4617): cast for X = 2 it asks for up to two bullets, for X = 6 up to
// all four, since a bullet can't be chosen twice (CR 700.2d). Cast for
// X = 0, or put onto the battlefield without being cast, it can choose
// none, and the trigger is removed (CR 700.2b). Choosing no bullet at
// all is a legal answer at any X, and removes it too.
//
// The bullets run in printed order (CR 608.2c). The rummage asks for
// its discard and the later bullets wait for the answer (#2789); it
// draws even from an empty hand. "Destroy target token" takes any
// token, yours included, and "each player" includes you.
//
// No simplification.
func init() {
	crew := WhenThisEnters("The Ruinous Wrecking Crew — choose up to X",
		func(*game.Game, *game.StackItem) error { return nil })
	crew.Modes = ChooseUpToX(
		ModeDoing("Discard a card, then draw a card.", nil, func(item *game.StackItem, ctx *Context, _ int) error {
			return ctx.Game.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
				Player: item.Controller,
				Source: item.SourceCardID,
				N:      1,
			}, theRuinousWreckingCrewDraw, nil)
		}),
		ModeDoing("Target opponent loses 2 life.", TargetPlayer("target opponent", Opponent()),
			func(_ *game.StackItem, ctx *Context, occ int) error {
				t, ok := ModeTarget(ctx, occ)
				if !ok || t.Kind != game.TargetPlayer {
					return nil
				}
				return GainLife{Player: t.ID, Amount: -2}.Apply(ctx)
			}),
		ModeDoing("Destroy target token.", TargetPermanent("target token", IsTokenPredicate()),
			DestroyTheModesTarget),
		ModeDoing("Each player sacrifices a creature of their choice.", nil,
			func(_ *game.StackItem, ctx *Context, _ int) error {
				return EachPlayerSacrifices{Match: Creature(), Label: "a creature"}.Apply(ctx)
			}),
	)
	Register(Spec{
		OracleID:                   theRuinousWreckingCrewOracle,
		Name:                       "The Ruinous Wrecking Crew",
		Completeness:               CompletenessFull,
		XMatters:                   true,
		EntersWithCountersFromCast: []game.EntryCountersFromCast{XCounters(game.CounterPlusOne)},
		Triggered:                  []game.TriggeredAbility{crew},
	})
}

// theRuinousWreckingCrewDraw is the rummage's "then draw a card", keyed
// (#4256): a table waiting on the discard is a restore point.
var theRuinousWreckingCrewDraw = game.RegisterDiscardThen("choose-cards/the-ruinous-wrecking-crew-draw", func(g *game.Game, r game.DiscardsDone) error {
	return g.DrawNForEffect(r.Player, 1)
})
