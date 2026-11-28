package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shimmering Efreet — Creature — Efreet {2}{U}, 2/2 (Visions, S82 older sets
// batch 791, #5631):
//
//	"Flying
//	 Phasing
//	 Whenever this creature phases in, target creature phases out. (It phases in
//	 before its controller untaps during their next untap step.)"
//
// Flying and phasing are printed keywords (CR 702.26a). The trigger watches this
// creature's own phase-in (CR 702.26d: it triggers though nothing enters), and its
// target creature, any controller's and possibly itself, is chosen as the trigger
// goes on the stack; a target no longer legal when it resolves does nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "34f52d00-301d-4065-a86b-8393b4c6e99d",
		Name:            "Shimmering Efreet",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", game.KeywordPhasing},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventPhaseIn, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Kind == game.EventPhaseIn && ev.CardID == source.InstanceID
			}, "Shimmering Efreet — target creature phases out", func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id := effects.FirstLegalCardTarget(ctx)
				if id == uuid.Nil {
					return nil
				}
				return effects.PhaseOut{Targets: []uuid.UUID{id}}.Apply(ctx)
			}), effects.TargetCreature("target creature")),
		},
	})
}
