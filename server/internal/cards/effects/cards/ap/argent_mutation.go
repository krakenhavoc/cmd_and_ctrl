package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Argent Mutation — Instant {2}{U}:
//
//	"Target permanent becomes an artifact in addition to its other types until
//	 end of turn.
//	 Draw a card."
//
// The card is drawn even when the target is gone by resolution only if the
// spell resolves at all: with its only target illegal it does nothing
// (CR 608.2b).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "652ef093-06e9-4bf1-b9cf-950fe043b319",
		Name:         "Argent Mutation",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Targets:      effects.TargetPermanent("target permanent"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.ScopedEffectFor{
					Target:   t.ID,
					Mods:     []game.Mod{game.AddTypesMod("Artifact")},
					Duration: effects.DurationUntilEndOfTurn(ctx),
					Label:    "Argent Mutation — becomes an artifact in addition to its other types",
				}).Apply(ctx); err != nil {
					return err
				}
			}
			if len(ctx.LegalTargets()) == 0 {
				return nil
			}
			return effects.DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
		},
	})
}
