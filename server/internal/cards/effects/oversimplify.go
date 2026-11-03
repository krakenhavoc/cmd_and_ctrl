package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Oversimplify — Sorcery {3}{G}{U} (Secrets of Strixhaven Commander,
// #4834):
//
//	"Exile all creatures. Each player creates a 0/0 green and blue
//	 Fractal creature token and puts a number of +1/+1 counters on it
//	 equal to the total power of creatures they controlled that were
//	 exiled this way."
//
// The exile is one simultaneous event (ExileAllMatching). Its tail is
// told the pre-move copies of what really reached exile (CR 603.10), so
// each creature's power is read as it last existed on the battlefield,
// counters and continuous effects included, under the controller it had
// then (the ruling). A commander in the sweep counts; its owner is asked
// about the command zone afterwards. A total below zero is zero (CR
// 107.1b).
//
// Then each player still in the game, in turn order from the active
// player, creates a Fractal and puts that many counters on it: the
// token enters 0/0 and the counters are put on it afterwards, as
// printed, so a player whose creatures had no power (or who had none)
// gets a 0/0 that dies to the state-based check. Every creation and every
// placement is keyed, so a replacement prompt along the way (a token or
// counter doubler) is a restore point; a doubled token gets the counters
// too.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "25e1d5fe-c2d8-4a11-ad54-bf1626f9ab45",
		Name:         "Oversimplify",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepExile}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return ExileAllMatching{Match: Creature(), ThenTail: ctx.ItemTail(oversimplifyExiled)}.Apply(ctx)
		},
	})
}

// oversimplifyExiled totals the exiled power per player and starts the
// Fractals. An on-disk key.
var oversimplifyExiled = MassTail("mass/oversimplify-exiled", func(ctx *Context, swept []game.Card, _ int) error {
	power := map[uuid.UUID]int{}
	for _, c := range swept {
		power[c.Controller] += c.PowerForComparison()
	}
	players := apnapPlayers(ctx.Game)
	totals := make([]int, len(players))
	for i, p := range players {
		totals[i] = max(power[p], 0)
	}
	return oversimplifyStep(ctx, nil, 0, players, totals)
})

// oversimplifyStep puts n counters on the next token still waiting for
// them, or, with none left, makes the next player's Fractal.
func oversimplifyStep(ctx *Context, pending []uuid.UUID, n int, players []uuid.UUID, totals []int) error {
	g := ctx.Game
	for len(pending) > 0 {
		tok := pending[0]
		pending = pending[1:]
		if n <= 0 || !onBattlefield(g, tok) {
			continue
		}
		c, ok := g.LookupCardForEffect(tok)
		if !ok {
			continue
		}
		then := ctx.ItemTail(oversimplifyCountersPut)
		then.Carry = packIDLists(pending, players)
		then.Nums = append([]int{n}, totals...)
		return g.AddCounterByKeyedForEffect(c.Controller, tok, game.CounterPlusOne, n, then)
	}
	if len(players) == 0 || len(totals) == 0 {
		return nil
	}
	then := ctx.ItemTail(oversimplifyFractalMade)
	then.N = totals[0]
	then.Carry = append([]uuid.UUID(nil), players[1:]...)
	then.Nums = append([]int(nil), totals[1:]...)
	return g.CreateTokensKeyedForEffect(game.TokenCreation{
		Controller: players[0],
		Source:     ctx.Source(),
		Groups: []game.TokenGroup{{
			Template: ctx.TextToken(TokenCard("0/0 green and blue Fractal")),
			Count:    1,
		}},
	}, then)
}

// oversimplifyFractalMade is told the Fractal (or Fractals) a player
// made; their counters come next. oversimplifyCountersPut continues once
// a Fractal's counters have settled. On-disk keys, registered in init:
// each runs the step that names it.
var oversimplifyFractalMade, oversimplifyCountersPut game.TailRef

func init() {
	oversimplifyFractalMade = CardTail("oversimplify/fractal-made", func(ctx *Context, o game.TailOutcome) error {
		return oversimplifyStep(ctx, o.IDs, o.Tail.N, o.Tail.Carry, o.Tail.Nums)
	})
	oversimplifyCountersPut = CardTail("oversimplify/counters-put", func(ctx *Context, o game.TailOutcome) error {
		pending, players := unpackIDPair(o.Tail.Carry)
		if len(o.Tail.Nums) == 0 {
			return nil
		}
		return oversimplifyStep(ctx, pending, o.Tail.Nums[0], players, o.Tail.Nums[1:])
	})
}
