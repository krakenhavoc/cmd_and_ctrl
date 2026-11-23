package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Censor — Instant {1}{U}
// (Amonkhet, #5535):
//
//	"Counter target spell unless its controller pays {1}.
//	 Cycling {U} ({U}, Discard this card: Draw a card.)"
//
// Force Spike's counter-unless-paid, plus cycling.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3a42558c-7432-4c05-9599-95343eb09be1",
		Name:         "Censor",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return effects.CounterUnlessPaid{
				StackID:  item.Targets[0].ID,
				Cost:     "{1}",
				Question: "Censor — pay {1} or your spell is countered",
			}.Apply(ctx)
		},
		Activated: []effects.ActivatedAbility{effects.Cycling("{U}")},
	})
}
