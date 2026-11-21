package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "f277b082-dfc8-45c6-bab5-b63cd218224c",
		Name:         "Silverchase Fox",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{W}, Sacrifice this creature: Exile target enchantment.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.Plus(effects.ManaCost("{1}{W}"), effects.SacrificeThis()),
			Targets: effects.TargetPermanent("target enchantment", effects.Enchantment()),
			Effect:  effects.ExileFirstLegalCardTarget,
		}},
	})
}
