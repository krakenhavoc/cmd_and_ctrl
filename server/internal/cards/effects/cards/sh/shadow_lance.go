package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shadow Lance — Enchantment — Aura {W} (Guildpact):
//
//	"Enchant creature
//	 Enchanted creature has first strike.
//	 {1}{B}: Enchanted creature gets +2/+2 until end of turn."
//
// Galvanic Arc's first strike grant, and Erebos's Emissary's pump of the
// creature the Aura enchants, read as the ability resolves (CR 608.2h): an
// Aura that has fallen off pumps nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f1a2a78c-2794-4bc9-9b06-49306378ef69",
		Name:         "Shadow Lance",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.GrantToAttached("first strike")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}: Enchanted creature gets +2/+2 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{1}{B}"),
			Effect:  effects.S82OldB377PumpSelfOrEnchanted("Shadow Lance — +2/+2", 2, 2),
		}},
	})
}
