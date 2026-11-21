package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Serpentine Ambush — Instant {1}{U} (Innistrad: Crimson Vow, S82):
//
//	"Until end of turn, target creature becomes a blue Serpent with base
//	 power and toughness 5/5."
//
// Polymorphist's Jest's layer mods on one creature: its creature subtypes
// become Serpent (CR 205.1a), its colour becomes blue (CR 105.3), and its
// base power and toughness are 5/5 (CR 613.4b); it keeps its abilities
// and counters and modifiers still apply on top. The set is locked as it
// resolves (CR 611.2c) and lasts until cleanup (CR 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "822b6d56-939a-42b9-b50c-9a8ba50ec3ad",
		Name:         "Serpentine Ambush",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := effects.UntilEndOfTurn(ctx, t.ID, nil, "Serpentine Ambush — a blue Serpent with base power and toughness 5/5",
					game.SetCreatureTypesMod(ctx.TextSubtypes("Serpent")...), game.SetColorsMod(ctx.TextColors("U")...),
					game.SetBasePowerMod(5), game.SetBaseToughnessMod(5)); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
