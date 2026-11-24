package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Suspend — Instant {U} (Modern Horizons 2, #5501):
//
//	"Exile target creature and put two time counters on it. If it doesn't
//	 have suspend, it gains suspend. (At the beginning of its owner's upkeep,
//	 they remove a time counter. When the last is removed, they may play it
//	 without paying its mana cost. If it's a creature, it has haste.)"
//
// Kang Prime's exile and suspend grant (CR 702.62a): the creature is exiled
// first and the counters and the granted suspend follow on the card that
// arrived, a creature that already has suspend getting only the counters.
// A target that is gone makes the spell do nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fef9425d-82d5-470b-9cbc-8c03a001099e",
		Name:         "Suspend",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return effects.ExileTarget{
				Target:   id,
				ThenTail: game.KeyedTail{Ref: game.GainSuspendExileTail, Player: ctx.Controller(), N: 2},
			}.Apply(ctx)
		},
	})
}
