package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Channel the Suns — Sorcery {3}{G} (Fifth Dawn, #5604):
//
//	"Add {W}{U}{B}{R}{G}."
//
// Seething Song's ritual, adding one mana of each colour to its caster's
// pool, which empties as the step ends (CR 500.4).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d1b1be96-b2aa-4386-97d2-eb0a8cc47210",
		Name:         "Channel the Suns",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.AddMana{Player: item.Controller, Produced: "{W}{U}{B}{R}{G}"}.Apply(ctx)
		},
	})
}
