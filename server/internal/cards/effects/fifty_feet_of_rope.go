package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:     "66aa6905-86f9-4800-988d-ffebafd4a9dc",
		Name:         "Fifty Feet of Rope",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "Climb Over — {T}: Target Wall can't block this turn.",
				Purpose: ForTargets(HarmUntilEOT(0, 0, 0)),
				Cost:    TapCost(),
				Targets: TargetPermanent("target Wall", Subtype("Wall")),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return RestrictUntilEOT{Target: id, Restrictions: game.CantBlock, Label: "Fifty Feet of Rope — can't block this turn"}.Apply(ctx)
				},
			},
			{
				Label:   "Tie Up — {3}, {T}: Target creature doesn't untap during its controller's next untap step.",
				Cost:    Plus(ManaCost("{3}"), TapCost()),
				Targets: TargetCreature("target creature"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return DoesntUntapNextUntapStep{Targets: []uuid.UUID{id}, Label: "Fifty Feet of Rope — doesn't untap during its controller's next untap step"}.Apply(ctx)
				},
			},
			{
				Label:        "Rappel Down — {4}, {T}: Venture into the dungeon. Activate only as a sorcery.",
				Cost:         Plus(ManaCost("{4}"), TapCost()),
				SorcerySpeed: true,
				Effect:       Venture,
			},
		},
	})
}
