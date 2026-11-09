package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Step Between Worlds — Sorcery {3}{U}{U} (Outlaws of Thunder Junction,
// #4468):
//
//	"Each player may shuffle their hand and graveyard into their library.
//	 Each player who does draws seven cards. Exile Step Between Worlds.
//	 Plot {4}{U}{U}"
//
// Each player is asked in turn order (CR 101.4), the caster first, and
// every answer is kept before anything moves: the players who said yes then
// shuffle their hand and graveyard in (Echo of Eons' tuck-and-shuffle), and
// once all those shuffles have landed each of them draws seven. The spell
// exiles itself as it resolves (Alchemist's Gambit). The question chain is
// plain data (a registered option-pick continuation carrying the players
// who said yes and those still to ask), so a table waiting on a player's
// answer is a restore point.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "48b43d53-9a8a-47a2-8610-a54d276d7be7",
		Name:           "Step Between Worlds",
		Completeness:   CompletenessFull,
		SpecialActions: []game.SpecialAction{Plot("{4}{U}{U}")},
		Purpose:        game.Purpose{Draws: 7},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := b35AskShuffleIn(ctx, nil, tablePlayers(ctx)); err != nil {
				return err
			}
			if sourceIsNewObject(ctx.Game, item) {
				return nil
			}
			return ctx.Game.ExileCardForEffect(ctx.Source())
		},
	})
}

const (
	stepBetweenWorldsYes = "Shuffle your hand and graveyard into your library, then draw seven cards"
	stepBetweenWorldsNo  = "Don't"
)

// b35AskShuffleIn asks the next player on `rest` whether they shuffle; with
// nobody left, the players who said yes shuffle in and draw.
func b35AskShuffleIn(ctx *Context, yes, rest []uuid.UUID) error {
	if len(rest) == 0 {
		return echoOfEonsShuffleEach(ctx.Game, yes, yes)
	}
	carry := append(append([]uuid.UUID(nil), yes...), rest[1:]...)
	return PickOption{
		Player:   rest[0],
		Question: "Step Between Worlds — shuffle your hand and graveyard into your library?",
		Options:  []game.ChoiceOption{{Label: stepBetweenWorldsNo}, {Label: stepBetweenWorldsYes}},
		ThenKey:  stepBetweenWorldsAnswer,
		Carry:    carry,
		Nums:     []int{len(yes)},
	}.Apply(ctx)
}

// stepBetweenWorldsAnswer records one player's answer and asks the next. Set
// in init, because its body names itself through the next question. An
// on-disk key: never renamed, never reused.
var stepBetweenWorldsAnswer game.OptionPickThen

func init() {
	stepBetweenWorldsAnswer = OptionPickThen("option-pick/step-between-worlds-shuffle", func(ctx *Context, r game.OptionPicked) error {
		n := 0
		if len(r.Nums) > 0 {
			n = min(max(r.Nums[0], 0), len(r.Carry))
		}
		yes := append([]uuid.UUID(nil), r.Carry[:n]...)
		rest := r.Carry[n:]
		if r.Option != nil && r.Option.Label == stepBetweenWorldsYes {
			yes = append(yes, r.Chooser)
		}
		return b35AskShuffleIn(ctx, yes, rest)
	})
}
