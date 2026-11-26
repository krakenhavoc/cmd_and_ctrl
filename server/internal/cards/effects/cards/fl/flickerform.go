package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flickerform — Enchantment — Aura {1}{W} (Commander 2013, S82 older sets):
//
//	"Enchant creature
//	 {2}{W}{W}: Exile enchanted creature and all Auras attached to it. At the
//	 beginning of the next end step, return that card to the battlefield
//	 under its owner's control. If you do, return the other cards exiled this
//	 way to the battlefield under their owners' control attached to that
//	 creature."
//
// Everything leaves in one exile, creature first, and Flickerform is among the
// Auras (it is attached to the creature). At the next end step the creature
// comes back as a new object under its owner's control; only if it did do the
// other cards return, together, each under its owner's control and attached to
// the new creature (CR 303.4f, an Aura that can't enchant it stays in exile,
// CR 303.4i). A token creature, or one that left exile first, doesn't return, so
// neither do the Auras.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e5345c28-7046-4ff3-a5d6-eeb7a0fb230b",
		Name:         "Flickerform",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{W}{W}: Exile enchanted creature and all Auras attached to it. At the beginning of the next end step, return that card to the battlefield under its owner's control. If you do, return the other cards exiled this way to the battlefield under their owners' control attached to that creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.ManaCost("{2}{W}{W}"),
			Effect:  effects.OldB383FlickerformExile,
		}},
	})
}
