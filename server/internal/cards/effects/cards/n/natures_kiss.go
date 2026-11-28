package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nature's Kiss — Enchantment — Aura {1}{G} (Weatherlight):
//
//	"Enchant creature
//	 {1}, Exile the top card of your graveyard: Enchanted creature gets +1/+1
//	 until end of turn."
//
// The cost exiles the topmost card of the activator's graveyard, whatever it is
// (CR 404.2, #6194). The ability is the Aura's controller's, whoever controls
// the creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e3e87fee-5717-4ca1-adfd-4df8732020f8",
		Name:         "Nature's Kiss",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, Exile the top card of your graveyard: Enchanted creature gets +1/+1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.ExileTopFromGraveyard(1, "the top card of your graveyard", nil)),
			Effect:  effects.OldB774EnchantedCreatureGetsUntilEOT(1, 1, "Nature's Kiss — +1/+1"),
		}},
	})
}
