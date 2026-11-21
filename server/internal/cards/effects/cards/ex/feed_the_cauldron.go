package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Feed the Cauldron — Instant {2}{B} (Wilds of Eldraine, #4471):
//
//	"Destroy target creature with mana value 3 or less. If it's your turn,
//	 create a Food token. (It's an artifact with "{2}, {T}, Sacrifice this
//	 token: You gain 3 life.")"
//
// The target's mana value is read where it is (CR 202.3), at cast and
// again as the spell resolves; a target that has become illegal counters
// the whole spell, Food included (CR 608.2b). "If it's your turn" is read
// as it resolves, so cast in an opponent's turn it only destroys. The
// Food is the catalog's token (CR 111.10b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "21153db6-552a-4956-a53a-a9dfb6968629",
		Name:         "Feed the Cauldron",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature with mana value 3 or less", effects.ManaValueLE(3)),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if t, ok := ctx.ClauseTarget(0); ok {
				if err := (effects.DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
			}
			if !effects.IsYourTurn(ctx.Game, item.Controller) {
				return nil
			}
			return effects.CreateToken{Template: effects.FoodToken(), N: 1}.Apply(ctx)
		},
	})
}
