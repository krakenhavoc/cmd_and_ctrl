package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func init() {
	Register(Spec{
		OracleID:      "8843faa9-0cd5-4ef3-8e7b-b8833db9dc84",
		Name:          "Rooftop Nuisance",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Casualty(1)},
		Targets:       TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (TapAndFreeze{Targets: []uuid.UUID{id}, Label: "Rooftop Nuisance — doesn't untap during its controller's next untap step"}).Apply(ctx); err != nil {
				return err
			}
			return DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
		},
	})
}
