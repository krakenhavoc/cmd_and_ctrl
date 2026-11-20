package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vampire Aristocrat — Creature — Vampire Rogue Noble {2}{B}, 2/2 (Magic 2010, #5585):
//
//	"Sacrifice a creature: This creature gets +2/+2 until end of turn."
//
// Bloodthrone Vampire's sacrifice outlet: the creature is sacrificed as the
// cost, and the pump lasts until end of turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c14739b0-5c19-4873-b027-314225a10f7b",
		Name:         "Vampire Aristocrat",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice a creature: This creature gets +2/+2 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerSacOutlet},
			Cost:    SacrificeACreature(),
			Effect:  thisGetsUntilEndOfTurn(2, 2, "Vampire Aristocrat — +2/+2 until end of turn"),
		}},
	})
}
