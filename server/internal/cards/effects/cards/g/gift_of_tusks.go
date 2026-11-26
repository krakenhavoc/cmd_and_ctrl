package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gift of Tusks — Instant {U} (Oath of the Gatewatch, S82 OLD #5545):
//
//	"Until end of turn, target creature loses all abilities and becomes a
//	 green Elephant with base power and toughness 3/3."
//
// Turn to Frog's one until-end-of-turn effect: parts in their layers at one
// timestamp (CR 611.2c) — the type change (layer 4), the colour (layer 5),
// the loss of abilities (layer 6) and the base 3/3 (layer 7b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c0307f8d-b39a-4180-8a45-ddc019ee39f6",
		Name:         "Gift of Tusks",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id := effects.FirstLegalBattlefieldTarget(ctx)
			if id == [16]byte{} {
				return nil
			}
			return effects.UntilEndOfTurn(ctx, id, nil, "Gift of Tusks — loses all abilities and becomes a green Elephant with base power and toughness 3/3",
				game.SetTypesMod([]string{"Creature"}, ctx.TextSubtypes("Elephant")),
				game.SetColorsMod(ctx.TextColor("G")),
				game.LoseAllAbilitiesMod(),
				game.SetBasePowerMod(3),
				game.SetBaseToughnessMod(3),
			)
		},
	})
}
