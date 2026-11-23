package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fierce Retribution — Instant {1}{W}:
//
//	"Cleave {5}{W} (You may cast this spell for its cleave cost. If you
//	 do, remove the words in square brackets.)
//	 Destroy target [attacking] creature."
//
// Hard-cast it needs a creature that is attacking as the spell is cast
// and as it resolves; cleaved any creature is legal (CR 702.148a).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "83094dd2-b8e9-4753-a8a0-657b446def60",
		Name:         "Fierce Retribution",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target attacking creature", effects.AttackingCreature()),
		AlternativeCosts: []game.AlternativeCost{
			effects.Cleave("{5}{W}", effects.TargetCreature("target creature")),
		},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			t, ok := effects.FirstTargetCard(item)
			if !ok {
				return nil
			}
			return effects.DestroyTarget{Target: t.ID}.Apply(ctx)
		},
	})
}
