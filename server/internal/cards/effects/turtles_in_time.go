package effects

import (
	"strconv"

	"github.com/google/uuid"

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
// The choice is Kwain, Itinerant Meddler's collector: each player is
// asked in APNAP order (CR 101.4), eliminated seats skipped, and nothing
// happens until the last answer. Only then do the players who said yes
// shuffle, Echo of Eons' two phases over just them: each one's hand and
// graveyard go into their library as one tuck and are shuffled, and once
// every shuffle has landed each of them draws seven. So nobody answers
// after seeing another player's new hand, and no draw happens between
// two answers. A "no" keeps that player's hand, creatures just returned
// to it included, and their graveyard. Turtles in Time is on the stack
// while all of this happens, so it is never shuffled away, and it goes
// to exile, not the graveyard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "803971e6-67f3-4fda-aeb6-69ef8ea4b54e",
		Name:         "Turtles in Time",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepBounce}},
		OnResolve: ResolveSteps(
			BounceAllMatching{Match: Creature()},
			Instruction(func(ctx *Context) error {
				return turtlesInTimeAsk(ctx, apnapPlayers(ctx.Game), nil)
			}),
			Instruction(func(ctx *Context) error {
				return ctx.Game.ExileCardForEffect(ctx.Item.SourceCardID)
			}),
		),
	})
}

// turtlesInTimeAsk asks the next player in `order`, carrying the players
// who have said yes so far. After the last answer the yes-sayers shuffle,
// then draw.
func turtlesInTimeAsk(ctx *Context, order, yes []uuid.UUID) error {
	if len(order) == 0 {
		if len(yes) == 0 {
			return nil
		}
		return echoOfEonsShuffleEach(ctx.Game, yes, yes)
	}
	who, rest := order[0], order[1:]
	if p := ctx.Game.PlayerByIDForEffect(who); p == nil || p.Eliminated {
		return turtlesInTimeAsk(ctx, rest, yes)
	}
	return MayChoice{
		Player:   who,
		Question: "Turtles in Time — shuffle your hand and graveyard into your library, then draw seven cards?",
		YesLabel: "Shuffle and draw seven",
		NoLabel:  "Keep them",
		Then:     turtlesInTimeAnswer,
		// The players who have said yes, then the ones still to ask.
		Carry: append(append([]uuid.UUID(nil), yes...), rest...),
		Word:  strconv.Itoa(len(yes)),
	}.Apply(ctx)
}

// turtlesInTimeAnswer is one player's answer: a yes is recorded, and
// either way the next player is asked. The prompt carries the yes-sayers
// (the first Word of them) and the players still to ask. Registered in
// init because it asks the next player through turtlesInTimeAsk, which
// names it. An on-disk key.
var turtlesInTimeAnswer game.ConfirmThen

func init() {
	turtlesInTimeAnswer = MayAnswer("confirm/turtles-in-time-ask", func(ctx *Context, r game.ConfirmAnswer) error {
		n := min(wordInt(r), len(r.Carry))
		yes, rest := append([]uuid.UUID(nil), r.Carry[:n]...), r.Carry[n:]
		if r.Accepted {
			yes = append(yes, r.Chooser)
		}
		return turtlesInTimeAsk(ctx, rest, yes)
	})
}
