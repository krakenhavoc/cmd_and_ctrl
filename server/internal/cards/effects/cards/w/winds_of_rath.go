package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Winds of Rath — Sorcery {3}{W}{W} (EDHREC rank 1647):
//
//	"Destroy all creatures that aren't enchanted. They can't be
//	 regenerated."
//
// The Aura deck's one-sided wrath. DestroyAllMatching over creatures
// with no Aura attached — an Equipment does not enchant, and an
// opponent's Aura DOES save its creature, as printed ("enchanted",
// not "enchanted by an Aura you control"). Attachment is #379's
// AttachedTo relation, read at resolution, so an Aura that fell off
// in response leaves its creature exposed. One simultaneous event,
// so a dies-payoff sees every creature that fell together.
//
// "They can't be regenerated" is ENFORCED as of #667
// (CR 701.19c) — the same rider Wrath of God carries.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID: "a6fd90dc-0ec3-4dce-a77a-4d04f5e254bf",
		Name:     "Winds of Rath",
		// ADR 0126 §6: an enchanted creature stays.
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DestroyAllMatching{Match: effects.And(effects.Creature(), effects.B15NotEnchanted()), CantBeRegenerated: true}.Apply(ctx)
		},
	})
}
