package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gravewaker — Creature — Bird Spirit {4}{B}{B}, 5/5 (Jumpstart, #5510):
//
//	"Flying
//	 {5}{B}{B}: Return target creature card from your graveyard to the
//	 battlefield tapped."
//
// Reassembling Skeleton's tapped entry on a targeted graveyard return.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7ceec1c7-c1f9-4d47-a203-6ee98c5d30d2",
		Name:            "Gravewaker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}{B}{B}: Return target creature card from your graveyard to the battlefield tapped.",
			Cost:    effects.ManaCost("{5}{B}{B}"),
			Targets: effects.TargetCardInGraveyard("target creature card from your graveyard", effects.Creature(), effects.YouOwn()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					return effects.ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield, Controller: ctx.Controller(), Tapped: true}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
