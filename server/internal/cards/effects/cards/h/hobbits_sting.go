package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hobbit's Sting — Instant {1}{W} (The Lord of the Rings: Tales of
// Middle-earth, #5948):
//
//	"Hobbit's Sting deals X damage to target creature, where X is the number
//	 of creatures you control plus the number of Foods you control."
//
// X is counted as the spell resolves, through the layers. The two counts are
// added, so a Food creature is counted once as a creature and once as a Food.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "845bbad1-02f9-432b-aca4-13a5ec71fd06",
		Name:         "Hobbit's Sting",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			x := 0
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.Controller != item.Controller {
					continue
				}
				if c.IsCreature() {
					x++
				}
				if c.HasSubtype("Food") {
					x++
				}
			}
			return effects.DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: x}.Apply(ctx)
		},
	})
}
