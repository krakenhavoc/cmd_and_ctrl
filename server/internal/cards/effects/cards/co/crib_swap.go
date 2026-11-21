package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crib Swap — Kindred Instant — Shapeshifter {2}{W} (Lorwyn Eclipsed,
// #4458; first printed in Lorwyn):
//
//	"Changeling (This card is every creature type.)
//	 Exile target creature. Its controller creates a 1/1 colorless
//	 Shapeshifter creature token with changeling."
//
// Resculpt's shape: the token goes to the exiled creature's controller,
// read before the exile, and "its controller" is honoured for your own
// creature as for an opponent's. With the target gone by resolution the
// spell does nothing and no token is made. Changeling is printed and
// honoured in every zone (Nameless Inversion); the token carries the
// keyword on its template (Springleaf Parade's 1/1 Shapeshifter).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2987c385-011a-4032-a516-a46d1e9dc9e8",
		Name:            "Crib Swap",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordChangeling},
		Targets:         effects.TargetCreature("target creature"),
		Purpose:         effects.ForTargets(effects.RemovalReturning(0, game.TargetReturn{CreatureTokens: 1, TokenPower: 1, TokenToughness: 1})),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			controller, ok := effects.ControllerOfTarget(ctx, target)
			if !ok {
				return nil
			}
			if err := (effects.ExileTarget{Target: target}).Apply(ctx); err != nil {
				return err
			}
			return effects.CreateToken{
				Controller: controller,
				Template:   effects.TokenCard("1/1 colorless Shapeshifter with changeling"),
				N:          1,
			}.Apply(ctx)
		},
	})
}
