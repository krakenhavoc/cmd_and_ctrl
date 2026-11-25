package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sleep of the Dead — Sorcery {U}:
//
//	"Tap target creature. It doesn't untap during its controller's next untap step.
//	 Escape—{2}{U}, Exile three other cards from your graveyard."
func init() {
	effects.Register(effects.Spec{
		OracleID:         "5e1b400f-ec5c-4703-bf30-65bb16bd3ca6",
		Name:             "Sleep of the Dead",
		Completeness:     effects.CompletenessFull,
		Targets:          effects.TargetCreature("target creature"),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Escape("{2}{U}", 3)},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return effects.TapAndFreeze{Targets: []uuid.UUID{id}, Label: "Sleep of the Dead — doesn't untap during its controller's next untap step"}.Apply(ctx)
		},
	})
}
