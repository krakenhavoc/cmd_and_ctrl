package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Courageous Resolve — Instant {2}{W} (Tales of Middle-earth Commander):
//
//	"Up to one target creature you control gains protection from each of
//	 your opponents until end of turn. Draw a card.
//	 Fateful hour — If you have 5 or less life, you can't lose life this
//	 turn, you can't lose the game this turn, and your opponents can't win
//	 the game this turn."
//
// The protection is Cliffside Rescuer's keyword grant, for the one creature
// still legal (the target is optional, so the draw happens either way, CR
// 608.2b). Fateful hour is a condition read as the spell resolves, after the
// draw, as printed: with 5 or less life you can't lose life this turn (a
// stored lock that stops a LOSS, a damage's life cost and a life payment, but
// not a gain; game.GrantLifeLossLockForEffect, CR 119.8), and Angel's
// Grace's game-end gates stop you losing and your opponents winning.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3f7452aa-b075-47d9-b61f-061d61c7d1b4",
		Name:         "Courageous Resolve",
		Completeness: effects.CompletenessFull,
		Purpose: game.Purpose{
			Draws:   1,
			Answers: game.AnswerProtect,
			Targets: game.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "protection")),
		},
		Targets: effects.TargetCreature("up to one target creature you control", effects.YouControl()).WithCount(0, 1),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.GrantKeywordUntilEOT{
					Target:   t.ID,
					Keywords: []string{game.ProtectionFromEachOfYourOpponents},
					Label:    "Courageous Resolve — protection from each of your opponents",
				}).Apply(ctx); err != nil {
					return err
				}
			}
			if err := (effects.DrawCards{N: 1}).Apply(ctx); err != nil {
				return err
			}
			p := ctx.PlayerByID(ctx.Controller())
			if p == nil || p.Life > 5 {
				return nil
			}
			ctx.Game.GrantLifeLossLockForEffect(ctx.Controller(), "Courageous Resolve — you can't lose life this turn", ctx.Source(), ctx.Game.UntilEndOfTurnDuration())
			return effects.CantLoseAndOpponentsCantWinThisTurn{Label: "Courageous Resolve"}.Apply(ctx)
		},
	})
}
