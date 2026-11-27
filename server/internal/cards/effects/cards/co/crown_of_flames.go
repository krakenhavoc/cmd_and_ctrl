package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crown of Flames — Enchantment — Aura {R} (Invasion, S82 older sets batch 661, #5616):
//
//	"Enchant creature
//	 {R}: Enchanted creature gets +1/+0 until end of turn.
//	 {R}: Return this Aura to its owner's hand."
//
// Firebreathing's pump with Conviction's return.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f14fb965-5ebc-427a-a5b8-149f54c64171",
		Name:         "Crown of Flames",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{R}: Enchanted creature gets +1/+0 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    effects.ManaCost("{R}"),
				Effect:  effects.S82OldB399PumpEnchanted("Crown of Flames", 1, 0),
			},
			{
				Label:   "{R}: Return this Aura to its owner's hand.",
				Purpose: game.Purpose{Answers: game.AnswerSurvive},
				Cost:    effects.ManaCost("{R}"),
				Effect:  effects.Do(effects.Instruction(effects.StdB9ReturnThisToHand)),
			},
		},
	})
}
