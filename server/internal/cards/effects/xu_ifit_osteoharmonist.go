package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Xu-Ifit, Osteoharmonist — Legendary Creature — Human Wizard {1}{B}{B},
// 2/3 (Edge of Eternities):
//
//	"{T}: Return target creature card from your graveyard to the
//	 battlefield. It's a Skeleton in addition to its other types and has
//	 no abilities. Activate only as a sorcery."
//
// The seam is game.EntersAs.LoseAllAbilities (ADR 0210's entry overwrite):
// the Skeleton type and the loss of every ability are one record
// registered as the card lands, before the zone move and EventETB, so
// the card's "when this enters" triggers, enters-tapped clauses and
// enters-with-counters clauses never apply (the rulings). It lasts as
// long as that object is on the battlefield; a flicker or a second
// reanimation is a new object. An ability granted to it afterwards is
// kept.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "39e32fdf-93bf-4390-8275-0d7d10ad7360",
		Name:         "Xu-Ifit, Osteoharmonist",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:        "{T}: Return target creature card from your graveyard to the battlefield. It's a Skeleton in addition to its other types and has no abilities. Activate only as a sorcery.",
			Cost:         TapCost(),
			SorcerySpeed: true,
			Targets:      TargetCardInGraveyard("target creature card from your graveyard", Creature(), YouOwn()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				targets := ctx.LegalTargets()
				if len(targets) == 0 {
					return nil
				}
				return putCardsFromZoneTogetherWith(ctx, game.ZoneGraveyard, []uuid.UUID{targets[0].ID}, game.ZoneEntryOptions{
					Controller: item.Controller,
					EntersAs: &game.EntersAs{
						AddSubtypes:      ctx.TextSubtypes("Skeleton"),
						LoseAllAbilities: true,
						Label:            "Xu-Ifit, Osteoharmonist — a Skeleton in addition to its other types, with no abilities",
					},
				}, game.TailRef{}, 0)
			},
		}},
	})
}
