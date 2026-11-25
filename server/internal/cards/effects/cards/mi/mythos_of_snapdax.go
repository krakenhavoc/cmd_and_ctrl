package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mythos of Snapdax — Sorcery {2}{W}{W} (Ikoria: Lair of Behemoths, #5512):
//
//	"Each player chooses an artifact, a creature, an enchantment, and a
//	 planeswalker from among the nonland permanents they control, then
//	 sacrifices the rest. If {B}{R} was spent to cast this spell, you choose
//	 the permanents for each player instead."
//
// Cataclysmic Gearhulk's resolution, with the chooser a parameter: each
// player's own pick, or, when both black and red mana paid for the spell,
// the caster's pick for every player's board. The mana spent is read as the
// spell resolves (ReadsOwnPayment).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "4b0b4d16-1d2b-497e-a1ae-bfb89eece665",
		Name:           "Mythos of Snapdax",
		Completeness:   effects.CompletenessFull,
		ReadsManaSpent: game.ReadsOwnPayment,
		WantsColors:    []string{"B", "R"},
		Purpose:        game.Purpose{Sweep: game.Sweep{Matches: game.SweepNonlandPermanents, How: game.SweepSacrifice, Partial: true}},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			youChoose := ctx.ManaSpentOfColor("B") > 0 && ctx.ManaSpentOfColor("R") > 0
			return effects.S82OldB119KeepOneOfEach(ctx.Game, item, youChoose)
		},
	})
}
