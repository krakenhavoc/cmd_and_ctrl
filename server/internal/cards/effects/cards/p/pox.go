package cardsp

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pox — Sorcery {B}{B}{B} (Mystery Booster 2, #5948):
//
//	"Each player loses a third of their life, then discards a third of the
//	 cards in their hand, then sacrifices a third of the creatures they
//	 control of their choice, then sacrifices a third of the lands they
//	 control of their choice. Round up each time."
//
// Pox Plague's four instructions in the printed order (CR 608.2c), each
// finishing before the next and each count worked out when its instruction
// starts, rounded up:
//
//   - life (caster first, then the opponents in seat order);
//   - cards in hand (DiscardPrompt.ThirdHandUp), chosen by the discarding
//     player, asked in APNAP order;
//   - creatures, chosen by their controller, APNAP, leaving together
//     (CR 101.4, 608.2e);
//   - lands, counted after the creatures have gone. A creature land is one
//     permanent and may be chosen for either step.
//
// The two sacrifices are registered continuations (poxSacrificeCreatures,
// poxSacrificeLands), so a table waiting on any seat's choice is a restore
// point.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "539d7c1a-b5e9-4204-9109-ddb2e8dec168",
		Name:         "Pox",
		Completeness: effects.CompletenessFull,
		OnResolve: effects.ResolveSteps(
			effects.Instruction(effects.GapB33EachPlayerLosesAThirdOfTheirLife),
			effects.Instruction(func(ctx *effects.Context) error {
				return ctx.Game.EachPlayerDiscardsThenKeyedForEffect(uuid.Nil,
					game.DiscardPrompt{
						Source:      ctx.Source(),
						ThirdHandUp: true,
						Question:    "Pox — discard a third of the cards in your hand, rounded up",
					},
					poxSacrificeCreatures, []uuid.UUID{ctx.Controller()})
			}),
		),
	})
}

// poxSacrificeCreatures is "sacrifices a third of the creatures they control
// of their choice", run once the last discard has landed.
var poxSacrificeCreatures = effects.RunThen("prompt-run/pox-sacrifice-creatures", func(ctx *effects.Context, _ game.RunDone) error {
	return effects.GapB33SacrificeAThird(ctx, effects.Creature(), "creatures you control", poxSacrificeLands)
})

// poxSacrificeLands is "then sacrifices a third of the lands they control of
// their choice", run once the creatures have left.
var poxSacrificeLands = effects.RunThen("prompt-run/pox-sacrifice-lands", func(ctx *effects.Context, _ game.RunDone) error {
	return effects.GapB33SacrificeAThird(ctx, effects.Land(), "lands you control", game.PromptRunThen{})
})
