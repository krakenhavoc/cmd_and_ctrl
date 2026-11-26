package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Drain the Well — Sorcery {2}{B/G}{B/G}:
//
//	"Destroy target land. You gain 2 life."
//
// Stone Rain's destroy, then the life. The life is gained even when the land
// is indestructible or regenerates, but not when the spell fizzles on an
// illegal target (CR 608.2b).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "84e89aac-439c-4e09-b7d9-2ee047c224c5",
		Name:         "Drain the Well",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target land", effects.Land()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.DestroyTheTargetPermanent(item, ctx); err != nil {
				return err
			}
			return effects.GainLife{Player: item.Controller, Amount: 2}.Apply(ctx)
		},
	})
}
