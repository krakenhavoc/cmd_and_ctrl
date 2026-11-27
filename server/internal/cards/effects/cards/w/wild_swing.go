package cardsw

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wild Swing — Sorcery {3}{R} (Shadowmoor, S82 older sets):
//
//	"Choose three target nonenchantment permanents. Destroy one of them at
//	 random."
//
// Capricious Efreet's destroy-one-target-at-random body over exactly three
// targets: of the targets still legal as the spell resolves, one is picked
// through the engine's keyed random stream and destroyed (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "92d07c7c-be30-4bcc-a632-99986fc5315c",
		Name:         "Wild Swing",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("three target nonenchantment permanents", effects.Not(effects.Enchantment())).WithCount(3, 3),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			var ids []uuid.UUID
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard && t.ID != uuid.Nil {
					ids = append(ids, t.ID)
				}
			}
			if len(ids) == 0 {
				return nil
			}
			pick := ctx.Game.ChooseAtRandomForEffect(effects.RandomDraw(ctx), ids, 1)
			if len(pick) == 0 {
				return nil
			}
			return effects.DestroyTarget{Target: pick[0]}.Apply(ctx)
		},
	})
}
