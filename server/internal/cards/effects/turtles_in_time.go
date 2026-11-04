package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Turtles in Time — Sorcery {5}{U}{U} (Teenage Mutant Ninja Turtles,
// #4457):
//
//	"Return all creatures to their owners' hands. Each player may
//	 shuffle their hand and graveyard into their library, then each
//	 player who does draws seven cards.
//	 Exile Turtles in Time."
//
// Three steps, each waiting for any question the one before asked (CR
// 608.2c): the mass bounce (a commander's owner may be asked about the
// command zone, CR 903.9b), then the choice, then the self-exile.
//
// The choice is Ruin Grinder's: every player, the caster included, is
// asked at once and answers for themselves. A "yes" is Echo of Eons'
// shuffle for that one player (their hand and graveyard into their
// library as one tuck, then a shuffle) followed by seven cards; a "no"
// keeps their hand, creatures just returned to it included, and their
// graveyard. Turtles in Time is still on the stack while players
// answer, so it is never shuffled away, and it goes to exile, not the
// graveyard, as it finishes resolving.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "803971e6-67f3-4fda-aeb6-69ef8ea4b54e",
		Name:         "Turtles in Time",
		Completeness: CompletenessFull,
		OnResolve: ResolveSteps(
			BounceAllMatching{Match: Creature()},
			Instruction(func(ctx *Context) error {
				for _, who := range tablePlayers(ctx) {
					if err := (MayChoice{
						Player:   who,
						Question: "Turtles in Time — shuffle your hand and graveyard into your library, then draw seven cards?",
						YesLabel: "Shuffle and draw seven",
						NoLabel:  "Keep them",
						Then:     turtlesInTimeAnswer,
					}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			}),
			Instruction(func(ctx *Context) error {
				return ctx.Game.ExileCardForEffect(ctx.Item.SourceCardID)
			}),
		),
	})
}

// turtlesInTimeAnswer is one player's "yes": their hand and graveyard go
// into their library together, and the shuffle and the draw follow once
// that tuck has landed. An on-disk key.
var turtlesInTimeAnswer = MayYes("confirm/turtles-in-time-shuffle", func(ctx *Context, r game.ConfirmAnswer) error {
	g := ctx.Game
	ids := append(allHandCardIDs(g, r.Chooser), echoOfEonsGraveyardCardIDs(g, r.Chooser)...)
	return g.TuckCardsToLibraryKeyedForEffect(ids, game.TuckOptions{},
		game.KeyedTail{Ref: turtlesInTimeShuffleThenDraw, Player: r.Chooser})
})

// turtlesInTimeShuffleThenDraw is that player's shuffle, then seven
// cards. An on-disk key.
var turtlesInTimeShuffleThenDraw = game.RegisterTail("then/turtles-in-time-shuffle-then-draw", func(g *game.Game, o game.TailOutcome) error {
	if err := g.ShuffleLibraryForEffect(o.Tail.Player); err != nil {
		return err
	}
	return g.DrawNForEffect(o.Tail.Player, 7)
})
