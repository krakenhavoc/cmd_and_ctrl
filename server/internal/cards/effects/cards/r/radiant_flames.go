package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Radiant Flames — Sorcery {2}{R}:
//
//	"Converge — Radiant Flames deals X damage to each creature, where X is
//	 the number of colors of mana spent to cast this spell."
//
// Converge is an ability word (CR 207.2c): the colours spent are read back at
// resolution from the stack item, as Arcane Omens does (WantsDistinctColors
// pays the generic half with colours not yet spent; ReadsManaSpent makes a
// waived payment ask what was paid). The damage is dealt to every creature
// as one simultaneous instance; with no colours counted it deals none.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:            "cfc8372f-13b6-4628-84c7-066881b287f5",
		Name:                "Radiant Flames",
		Completeness:        effects.CompletenessFull,
		ReadsManaSpent:      game.ReadsOwnPayment,
		WantsDistinctColors: true,
		Purpose:             game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, AmountIsX: true}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DamageEachMatching(ctx, effects.Creature(), ctx.ColorsSpentCount())
		},
	})
}
