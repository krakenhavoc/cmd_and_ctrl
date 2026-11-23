package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dispel — Instant, {U} (EDHREC rank 926):
//
//	"Counter target instant spell."
//
// Negate's narrower cousin: the clause admits instants only, so a
// sorcery, a creature or an ability is refused at announce.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6d7be242-a072-40ce-b540-95880506cccd",
		Name:         "Dispel",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target instant spell", effects.Instant()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			return effects.CounterTarget{StackID: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
