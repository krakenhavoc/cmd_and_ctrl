package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sequence Engine — Artifact {2}{G} (Commander 2021, #5504):
//
//	"{X}, {T}: Exile target creature card with mana value X from a
//	 graveyard. Create a 0/0 green and blue Fractal creature token. Put X
//	 +1/+1 counters on it. Activate only as a sorcery."
//
// X is announced with the activation (CR 107.3a) and the target clause reads
// it (WithManaValueEqualsX), checked as the ability is put on the stack and
// again as it resolves (CR 608.2b): a card whose mana value is not X, or one
// that has left, makes the ability do nothing. The Fractal is the
// create-then-put shape (SocB8FractalWithCountersSteps) with X counters.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "87c1bd51-6a85-4a14-af7b-a5b02ab69d73",
		Name:         "Sequence Engine",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Activated: []effects.ActivatedAbility{{
			Label:        "{X}, {T}: Exile target creature card with mana value X from a graveyard. Create a 0/0 green and blue Fractal creature token. Put X +1/+1 counters on it. Activate only as a sorcery.",
			Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
			Cost:         effects.Plus(effects.ManaCost("{X}"), effects.TapCost()),
			Targets:      effects.TargetCardInGraveyard("target creature card with mana value X from a graveyard", effects.Creature()).WithManaValueEqualsX(),
			SorcerySpeed: true,
			Effect: effects.Do(append([]effects.Applier{effects.Instruction(func(ctx *effects.Context) error {
				return effects.ExileFirstTarget(ctx.Game, ctx.Item)
			})}, effects.SocB8FractalWithCountersSteps("sequence-engine/fractal", func(ctx *effects.Context) int { return ctx.X() })...)...),
		}},
	})
}
