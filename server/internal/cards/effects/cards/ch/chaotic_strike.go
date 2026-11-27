package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chaotic Strike — Instant {1}{R} (Invasion, S82 #5616):
//
//	"Cast this spell only during combat after blockers are declared.
//	 Flip a coin. If you win the flip, target creature gets +1/+1 until end of
//	 turn.
//	 Draw a card."
//
// The timing is Curtain of Light's: nobody has priority in the declare
// blockers step until every blocker is declared, so each step from there to
// the end of combat is "after blockers are declared". The flip is the engine's
// called flip; the pump (if won) and the draw both run once it settles.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:           "1f0b3549-2d24-49e8-85bc-55e3a5e5f6b8",
		Name:               "Chaotic Strike",
		Completeness:       effects.CompletenessFull,
		Purpose:            game.Purpose{Draws: 1, Targets: effects.ForTargets(effects.BenefitUntilEOT(0, 1, 1)).Targets},
		CastCondition:      effects.DuringCombatAfterBlockers,
		CastConditionLabel: "Cast this spell only during combat after blockers are declared.",
		Targets:            effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			ctx.Game.FlipCoinForEffect(game.CoinFlipSpec{
				Flipper:  item.Controller,
				Source:   item.SourceCardID,
				Question: "Chaotic Strike — call the coin flip",
				ThenKey:  chaoticStrikeFlipped,
				Asked:    game.NewAskedItem(item),
			})
			return nil
		},
	})
}

// chaoticStrikeFlipped is "If you win the flip, target creature gets +1/+1
// until end of turn. Draw a card." An on-disk key.
var chaoticStrikeFlipped = effects.CoinFlipThen("coin-flip/chaotic-strike", func(ctx *effects.Context, r game.CoinFlipped) error {
	if effects.WonFirstFlip(r) {
		if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetCard {
			if err := (effects.BoostUntilEOT{Target: t.ID, Power: 1, Toughness: 1, Label: "Chaotic Strike — +1/+1"}).Apply(ctx); err != nil {
				return err
			}
		}
	}
	return effects.DrawCards{N: 1, Player: r.Flipper}.Apply(ctx)
})
