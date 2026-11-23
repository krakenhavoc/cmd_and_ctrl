package cardsex

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fifty Feet of Rope — Artifact {1}:
//
//	"Climb Over — {T}: Target Wall can't block this turn.
//	 Tie Up — {3}, {T}: Target creature doesn't untap during its
//	 controller's next untap step.
//	 Rappel Down — {4}, {T}: Venture into the dungeon. Activate only as
//	 a sorcery."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "66aa6905-86f9-4800-988d-ffebafd4a9dc",
		Name:         "Fifty Feet of Rope",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "Climb Over — {T}: Target Wall can't block this turn.",
				Purpose: effects.ForTargets(effects.CantBlockUntilEOT(0, 0, 0)),
				Cost:    effects.TapCost(),
				Targets: effects.TargetPermanent("target Wall", effects.Subtype("Wall")),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return effects.RestrictUntilEOT{Target: id, Restrictions: game.CantBlock, Label: "Fifty Feet of Rope — can't block this turn"}.Apply(ctx)
				},
			},
			{
				Label:   "Tie Up — {3}, {T}: Target creature doesn't untap during its controller's next untap step.",
				Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
				Targets: effects.TargetCreature("target creature"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return effects.DoesntUntapNextUntapStep{Targets: []uuid.UUID{id}, Label: "Fifty Feet of Rope — doesn't untap during its controller's next untap step"}.Apply(ctx)
				},
			},
			{
				Label:        "Rappel Down — {4}, {T}: Venture into the dungeon. Activate only as a sorcery.",
				Cost:         effects.Plus(effects.ManaCost("{4}"), effects.TapCost()),
				SorcerySpeed: true,
				Effect:       effects.Venture,
			},
		},
	})
}
