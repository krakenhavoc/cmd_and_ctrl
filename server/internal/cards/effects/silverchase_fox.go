package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Silverchase Fox — Creature — Fox {1}{W}, 2/2 (Innistrad, #5572):
//
//	"{1}{W}, Sacrifice this creature: Exile target enchantment."
//
// Lawbringer's sacrifice-and-exile with Elf Replica's {1}{W} cost and
// enchantment target: the sacrifice is a cost, so the Fox is gone before
// the ability resolves (CR 602.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f277b082-dfc8-45c6-bab5-b63cd218224c",
		Name:         "Silverchase Fox",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{W}, Sacrifice this creature: Exile target enchantment.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    Plus(ManaCost("{1}{W}"), SacrificeThis()),
			Targets: TargetPermanent("target enchantment", Enchantment()),
			Effect:  exileFirstLegalCardTarget,
		}},
	})
}
