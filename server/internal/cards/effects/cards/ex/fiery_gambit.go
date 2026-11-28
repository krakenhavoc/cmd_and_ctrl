package cardsex

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fiery Gambit — Sorcery {2}{R} (Mirrodin, S82 OLD, #5606):
//
//	"Flip a coin until you lose a flip or choose to stop flipping. If you
//	 lose a flip, Fiery Gambit has no effect. If you win one or more flips,
//	 Fiery Gambit deals 3 damage to target creature. If you win two or more
//	 flips, Fiery Gambit deals 6 damage to each opponent. If you win three
//	 or more flips, draw nine cards and untap all lands you control."
//
// Mirror March's flip-until-you-lose chain with a stop: each link is one
// coin the controller calls (CR 705); a lost flip ends it with nothing, and
// each won flip asks whether to flip again. Declining settles on the flips
// won. The clauses stack, so three wins deal 3 to the creature, 6 to each
// opponent, then draw nine and untap the lands. The creature is the
// spell's one target, judged as it resolves (CR 608.2b): a target that is
// gone makes the whole spell fizzle before any flip, as printed. The count
// of wins rides the chain as data (the prompt's Word), so a table waiting
// on a call or on the question is a restore point.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "78497c54-7179-41b7-9454-7b915fd96cd5",
		Name:         "Fiery Gambit",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			fieryGambitFlip(ctx.Game, item, 0)
			return nil
		},
	})
}

// fieryGambitFlip queues one coin of the chain, `wins` flips won so far.
func fieryGambitFlip(g *game.Game, item *game.StackItem, wins int) {
	g.FlipCoinForEffect(game.CoinFlipSpec{
		Flipper:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Fiery Gambit — call the coin flip (flip until you lose or stop)",
		Wins:     wins,
		ThenKey:  fieryGambitFlipped,
		Asked:    game.NewAskedItem(item),
	})
}

// fieryGambitFlipped is the settled call: a lost flip ends the spell, a
// won one asks whether to flip again. An on-disk key, registered in init
// because its body queues the next link.
var fieryGambitFlipped game.CoinFlipThen

func init() {
	fieryGambitFlipped = effects.CoinFlipThen("coin-flip/s82-old-b602-fiery-gambit-flip", func(ctx *effects.Context, r game.CoinFlipped) error {
		if !effects.WonFirstFlip(r) {
			return nil
		}
		wins := r.Wins + 1
		return effects.MayChoice{
			Question: "Fiery Gambit — you have won " + strconv.Itoa(wins) + " flip(s). Flip again?",
			YesLabel: "Flip again",
			NoLabel:  "Stop flipping",
			Then:     fieryGambitAgain,
			Word:     strconv.Itoa(wins),
		}.Apply(ctx)
	})
}

// fieryGambitAgain is the answer to "flip again?": yes queues the next
// coin, no settles on the flips won (the answer's Word). An on-disk key.
var fieryGambitAgain = effects.MayAnswer("confirm/s82-old-b602-fiery-gambit-again", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	wins, err := strconv.Atoi(r.Word)
	if err != nil {
		return nil
	}
	if r.Accepted {
		fieryGambitFlip(ctx.Game, ctx.Item, wins)
		return nil
	}
	return fieryGambitSettle(ctx, wins)
})

// fieryGambitSettle is "if you win one or more flips … two or more …
// three or more": each clause in order, over the spell's own target.
func fieryGambitSettle(ctx *effects.Context, wins int) error {
	if wins >= 1 {
		if id, ok := effects.B16FirstLegalTargetCard(ctx); ok {
			if err := (effects.DealDamage{Source: ctx.Source(), Target: id, Amount: 3}).Apply(ctx); err != nil {
				return err
			}
		}
	}
	if wins >= 2 {
		if err := (effects.DamageEachOpponent{N: 6}).Apply(ctx); err != nil {
			return err
		}
	}
	if wins >= 3 {
		if err := (effects.DrawCards{N: 9}).Apply(ctx); err != nil {
			return err
		}
		return effects.B16UntapAllYouControlMatching(ctx, ctx.Controller(), func(c game.Card) bool { return c.IsLand() })
	}
	return nil
}
