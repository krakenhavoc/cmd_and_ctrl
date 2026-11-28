package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Crooked Scales — Artifact {4} (Mercadian Masques, S82 OLD, #5619):
//
//	"{4}, {T}: Flip a coin. If you win the flip, destroy target creature an
//	 opponent controls. If you lose the flip, destroy target creature you
//	 control unless you pay {3} and repeat this process."
//
// Both creatures are targets, chosen as the ability goes on the stack (CR 115.1),
// one an opponent's and one yours. The flip is Mogg Assassin's keyed call: won,
// the opponent's creature is destroyed; lost, the controller is asked whether to
// pay {3}. Paying spares their own creature and flips again, the same two targets
// standing, so the process repeats for as long as they keep paying and losing;
// declining destroys their own creature. A target that is no longer legal as the
// destruction happens is not destroyed (CR 608.2b).
//
// No simplification.
func init() {
	crookedScalesFlipped = effects.CoinFlipThen("coin-flip/crooked-scales", func(ctx *effects.Context, r game.CoinFlipped) error {
		if effects.WonFirstFlip(r) {
			return crookedScalesDestroy(ctx, 0)
		}
		return effects.MayPay{
			Chooser:  ctx.Controller(),
			Cost:     "{3}",
			Question: "Crooked Scales — pay {3} to spare your creature and flip again?",
			Then:     crookedScalesLost,
		}.Apply(ctx)
	})
	crookedScalesLost = effects.PayUnlessBranches("pay-unless/crooked-scales-lost",
		func(ctx *effects.Context, _ game.PayUnlessAnswer) error { return crookedScalesFlip(ctx) },
		func(ctx *effects.Context, _ game.PayUnlessAnswer) error { return crookedScalesDestroy(ctx, 1) })
	effects.Register(effects.Spec{
		OracleID:     "49c47768-fcc7-4091-b06d-d8a2bd7d1cc9",
		Name:         "Crooked Scales",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label: "{4}, {T}: Flip a coin. If you win the flip, destroy target creature an opponent controls. " +
				"If you lose the flip, destroy target creature you control unless you pay {3} and repeat this process.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.Plus(effects.ManaCost("{4}"), effects.TapCost()),
			Targets: effects.Clauses(
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
				effects.TargetCreature("target creature you control", effects.YouControl()),
			),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return crookedScalesFlip(effects.NewContext(g, item))
			},
		}},
	})
}

var (
	crookedScalesFlipped game.CoinFlipThen
	crookedScalesLost    game.PayUnlessElse
)

// crookedScalesFlip asks the controller to call a flip for the ability's item.
func crookedScalesFlip(ctx *effects.Context) error {
	ctx.Game.FlipCoinForEffect(game.CoinFlipSpec{
		Flipper:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: "Crooked Scales — call the coin flip",
		ThenKey:  crookedScalesFlipped,
		Asked:    game.NewAskedItem(ctx.Item),
	})
	return nil
}

// crookedScalesDestroy destroys the creature that answered clause `slot`, if it
// is still a legal target.
func crookedScalesDestroy(ctx *effects.Context, slot int) error {
	t, ok := ctx.ClauseTarget(slot)
	if !ok || t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
		return nil
	}
	return effects.DestroyTarget{Target: t.ID}.Apply(ctx)
}
