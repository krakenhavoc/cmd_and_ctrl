package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lithophage — Creature — Insect {3}{R}{R}, 7/7 (Mercadian Masques, S82 OLD batch 688, #5619):
//
//	"At the beginning of your upkeep, sacrifice this creature unless you sacrifice a Mountain."
//
// The Gitrog Monster's upkeep tax over Mountains (SacrificeThisUnlessYouSacrificeA): a two-way
// choice when its controller has a Mountain to give, and an unconditional sacrifice of the
// Lithophage when they have none. It is not a mana payment, so it is a keyed option pick.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b4eb3d7e-a234-4ed5-8611-fcb818686fcf",
		Name:         "Lithophage",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Lithophage — sacrifice this creature unless you sacrifice a Mountain",
				effects.SacrificeThisUnlessYouSacrificeA("Lithophage", "a Mountain", effects.HasSubtype("Mountain"), effects.OldB688SacrificeAMountainOrThisAnswer)),
		},
	})
}
