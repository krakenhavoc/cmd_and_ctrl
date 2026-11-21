package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Turn to Frog — Instant {1}{U} (Ravnica: Clue Edition):
//
//	"Until end of turn, target creature loses all abilities and becomes a
//	 blue Frog with base power and toughness 1/1."
//
// One until-end-of-turn effect locked as the spell resolves (CR 611.2c),
// its parts in their layers at one timestamp: the creature loses its
// abilities (layer 6), becomes a blue Frog (a type-setting change that
// keeps it a creature, layer 4, and a colour set, layer 5) and has base
// power and toughness 1/1 (layer 7b). Counters and +N/+N effects still
// apply on top. Amphibian Downpour's body for a turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "74c4c1e2-c50d-4c8f-889a-0d5674dc6d67",
		Name:         "Turn to Frog",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id := effects.FirstLegalBattlefieldTarget(ctx)
			if id == [16]byte{} {
				return nil
			}
			return effects.UntilEndOfTurn(ctx, id, nil, "Turn to Frog — loses all abilities and becomes a blue Frog with base power and toughness 1/1",
				game.SetTypesMod([]string{"Creature"}, ctx.TextSubtypes("Frog")),
				game.SetColorsMod(ctx.TextColor("U")),
				game.LoseAllAbilitiesMod(),
				game.SetBasePowerMod(1),
				game.SetBaseToughnessMod(1),
			)
		},
	})
}
