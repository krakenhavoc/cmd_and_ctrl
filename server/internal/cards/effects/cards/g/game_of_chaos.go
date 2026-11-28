package cardsg

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Game of Chaos — Sorcery {R}{R}{R} (Ice Age, S82 OLD batch 851, #5636):
//
//	"Flip a coin. If you win the flip, you gain 1 life and target opponent
//	 loses 1 life, and you decide whether to flip again. If you lose the flip,
//	 you lose 1 life and that opponent gains 1 life, and that player decides
//	 whether to flip again. Double the life stakes with each flip."
//
// Fiery Gambit's chain: each link is one coin the spell's controller calls
// (CR 705), settled by a registered key, with the flips so far riding the
// chain as data (the coin's Wins, the question's Word), so a table waiting on a
// call or on the question is a restore point. The stakes of a flip are 2 to the
// power of the flips already made (1, 2, 4, ...). A won flip moves the stakes in
// the controller's favour and asks the controller whether to flip again; a lost
// flip moves them the other way and asks the targeted opponent. Whoever is asked,
// the controller flips. The opponent is the spell's one target, judged as it
// resolves (CR 608.2b): a target that is gone makes the spell do nothing, and
// once the chain has started it carries on with the player it began with.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b6756608-a618-4e26-a090-ea8fc992def1",
		Name:         "Game of Chaos",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if effects.FirstLegalTargetPlayer(ctx) == uuid.Nil {
				return nil
			}
			gameOfChaosFlip(ctx.Game, item, 0)
			return nil
		},
	})
}

// gameOfChaosFlip queues one coin of the chain, `flips` flips made so far.
func gameOfChaosFlip(g *game.Game, item *game.StackItem, flips int) {
	g.FlipCoinForEffect(game.CoinFlipSpec{
		Flipper:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Game of Chaos — call the coin flip (" + strconv.Itoa(1<<flips) + " life at stake)",
		Wins:     flips,
		ThenKey:  gameOfChaosFlipped,
		Asked:    game.NewAskedItem(item),
	})
}

// gameOfChaosOpponent is the spell's targeted opponent, read from its
// announced targets (the chain outlives the legality check made as it began).
func gameOfChaosOpponent(item *game.StackItem) uuid.UUID {
	for _, t := range item.Targets {
		if t.Kind == game.TargetPlayer {
			return t.ID
		}
	}
	return uuid.Nil
}

// gameOfChaosFlipped is the settled call: the stakes change hands, then the
// winner of the flip's side decides whether to flip again. An on-disk key,
// registered in init because its body queues the next link.
var gameOfChaosFlipped game.CoinFlipThen

func init() {
	gameOfChaosFlipped = effects.CoinFlipThen("coin-flip/s82-old-b851-game-of-chaos-flip", func(ctx *effects.Context, r game.CoinFlipped) error {
		stakes := 1 << r.Wins
		me := ctx.Controller()
		opp := gameOfChaosOpponent(ctx.Item)
		src := ctx.Source()
		won := effects.WonFirstFlip(r)
		gain, lose := me, opp
		decider := me
		if !won {
			gain, lose = opp, me
			decider = opp
		}
		if err := ctx.Game.ChangePlayerLifeForEffect(src, gain, stakes); err != nil {
			return err
		}
		if err := ctx.Game.ChangePlayerLifeForEffect(src, lose, -stakes); err != nil {
			return err
		}
		if decider == uuid.Nil {
			return nil
		}
		return effects.MayChoice{
			Player:   decider,
			Question: "Game of Chaos — flip again? (" + strconv.Itoa(stakes*2) + " life at stake)",
			YesLabel: "Flip again",
			NoLabel:  "Stop flipping",
			Then:     gameOfChaosAgain,
			Word:     strconv.Itoa(r.Wins + 1),
		}.Apply(ctx)
	})
}

// gameOfChaosAgain is the answer to "flip again?": yes queues the next coin
// (the spell's controller flips), no ends the chain. An on-disk key.
var gameOfChaosAgain = effects.MayAnswer("confirm/s82-old-b851-game-of-chaos-again", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	flips, err := strconv.Atoi(r.Word)
	if err != nil || !r.Accepted {
		return nil
	}
	gameOfChaosFlip(ctx.Game, ctx.Item, flips)
	return nil
})
