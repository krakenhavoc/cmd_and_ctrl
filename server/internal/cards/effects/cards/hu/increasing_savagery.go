package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Increasing Savagery — Sorcery {2}{G}{G}:
//
//	"Put five +1/+1 counters on target creature. If this spell was cast from
//	 a graveyard, put ten +1/+1 counters on that creature instead.
//	 Flashback {5}{G}{G}"
func init() {
	effects.Register(effects.Spec{
		OracleID:         "616c7f45-5295-4b8c-a828-ff3a5ba4d917",
		Name:             "Increasing Savagery",
		Completeness:     effects.CompletenessFull,
		Targets:          effects.TargetCreature("target creature"),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{5}{G}{G}")},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			n := 5
			if item.CastFromZone == game.ZoneGraveyard {
				n = 10
			}
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				return effects.AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: n}.Apply(ctx)
			}
			return nil
		},
	})
}
