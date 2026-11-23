package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tenacity — Instant {3}{W} (Foundations Jumpstart, #5929):
//
//	"Creatures you control get +1/+1 and gain lifelink until end of
//	 turn. Untap those creatures."
//
// The set of creatures is fixed once, as the spell resolves (CR 611.2c),
// and "those creatures" are exactly the creatures you control at that
// moment, so the untap runs over the same set.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7dfe701f-6572-4733-9da5-09397600bba7",
		Name:         "Tenacity",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerCombatGrant | game.AnswerPump},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			yours := effects.And(effects.Creature(), effects.YouControl())
			if err := (effects.BoostUntilEOT{Match: yours, Power: 1, Toughness: 1, Label: "Tenacity — +1/+1"}).Apply(ctx); err != nil {
				return err
			}
			if err := (effects.GrantKeywordUntilEOT{Match: yours, Keywords: []string{"lifelink"}, Label: "Tenacity — lifelink"}).Apply(ctx); err != nil {
				return err
			}
			return effects.UntapAllCreaturesYouControl{}.Apply(ctx)
		},
	})
}
