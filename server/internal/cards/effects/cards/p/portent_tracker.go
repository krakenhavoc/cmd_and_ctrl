package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Portent Tracker — Creature — Satyr Scout {1}{G}, 1/1 (March of the
// Machine):
//
//	"{T}: Untap target land.
//	 {T}: Choose target battle. If an opponent protects it, remove a
//	 defense counter from it. Otherwise, put a defense counter on it.
//	 Activate only as a sorcery."
//
// Two separate abilities, each with its own {T}. The protector is read as
// the second resolves (CR 310.9): a battle with no protector counts as one
// an opponent does not protect.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ca49d6ad-97b1-4fd4-b1cf-fecd9cafee6d",
		Name:         "Portent Tracker",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{T}: Untap target land.",
				Cost:    effects.TapCost(),
				Targets: effects.TargetPermanent("target land", effects.Land()),
				Effect:  effects.B22UntapEachLegalTarget,
			},
			{
				Label:        "{T}: Choose target battle. If an opponent protects it, remove a defense counter from it. Otherwise, put a defense counter on it. Activate only as a sorcery.",
				Cost:         effects.TapCost(),
				Targets:      effects.S82MomB12TargetBattle(),
				SorcerySpeed: true,
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						return effects.S82MomB12DefenseEither(ctx, t.ID, 1)
					}
					return nil
				},
			},
		},
	})
}
