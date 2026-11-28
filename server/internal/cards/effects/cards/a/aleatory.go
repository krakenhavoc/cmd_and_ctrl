package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Aleatory — Instant {1}{R} (Mirage):
//
//	"Cast this spell only during combat after blockers are declared.
//	 Flip a coin. If you win the flip, target creature gets +1/+1 until end
//	 of turn.
//	 Draw a card at the beginning of the next turn's upkeep."
//
// The timing is every step from the declare blockers step through end of
// combat (nobody has priority in the declare blockers step until blockers are
// declared, #1501). The target is chosen when the spell is cast; the keyed
// coin flip runs while the spell resolves (Boompile's shape) and a won flip
// pumps the target if it is still legal (CR 608.2b). The delayed draw (CR
// 603.7) is scheduled as the spell resolves, win or lose.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "da4cab2d-0068-440f-ba7c-f2d4e5fa976e",
		Name:         "Aleatory",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1, Targets: game.ForTargets(effects.BenefitUntilEOT(0, 1, 1))},
		CastCondition: func(g *game.Game, player uuid.UUID, c game.Card) bool {
			return effects.OldB681AfterBlockersAreDeclared(g, player, c.InstanceID)
		},
		CastConditionLabel: "Cast this spell only during combat after blockers are declared.",
		Targets:            effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.DrawACardAtTheNextUpkeep("Aleatory — draw a card").Apply(ctx); err != nil {
				return err
			}
			ctx.Game.FlipCoinForEffect(game.CoinFlipSpec{
				Flipper:  item.Controller,
				Source:   item.SourceCardID,
				Question: "Aleatory — call the coin flip",
				ThenKey:  aleatoryFlipped,
				Asked:    game.NewAskedItem(item),
			})
			return nil
		},
	})
}

// aleatoryFlipped is "if you win the flip, target creature gets +1/+1 until
// end of turn". An on-disk key.
var aleatoryFlipped = effects.CoinFlipThen("coin-flip/aleatory", func(ctx *effects.Context, r game.CoinFlipped) error {
	if !effects.WonFirstFlip(r) {
		return nil
	}
	t, ok := ctx.ClauseTarget(0)
	if !ok {
		return nil
	}
	return effects.BoostUntilEOT{Target: t.ID, Power: 1, Toughness: 1, Label: "Aleatory — +1/+1"}.Apply(ctx)
})
