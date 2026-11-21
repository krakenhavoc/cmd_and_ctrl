package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Zhalfirin Shapecraft — Instant {1}{U} (March of the Machine, S82 #5477):
//
//	"Target creature has base power and toughness 4/3 until end of turn.
//	 Draw a card."
//
// Base power and toughness is a layer 7b effect (CR 613.4b), so counters,
// anthems and other pumps still apply on top of 4/3. It is locked to the
// target at resolution (CR 611.2c) and ends in the cleanup step (CR 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cf459f47-e565-4bbb-89e4-ed54da8adae1",
		Name:         "Zhalfirin Shapecraft",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.ScopedEffectFor{
					Target:   t.ID,
					Mods:     game.SetBasePTMods(4, 3),
					Duration: effects.DurationUntilEndOfTurn(ctx),
					Label:    "Zhalfirin Shapecraft — base power and toughness 4/3",
				}).Apply(ctx); err != nil {
					return err
				}
				break
			}
			return effects.DrawCards{N: 1}.Apply(ctx)
		},
	})
}
