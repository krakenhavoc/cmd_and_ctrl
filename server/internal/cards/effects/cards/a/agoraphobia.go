package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Agoraphobia — Enchantment — Aura {1}{U} (Gatecrash, S82 OLD b407, #5565):
//
//	"Enchant creature
//	 Enchanted creature gets -5/-0.
//	 {2}{U}: Return this Aura to its owner's hand."
//
// Cage of Hands' return row (StdB9ReturnThisToHand) with a -5/-0 layer 7c pump.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b268fe15-2860-4730-865e-eed7f308fc45",
		Name:         "Agoraphobia",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static:       []game.StaticAbility{effects.PumpAttached(-5, 0)},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{U}: Return this Aura to its owner's hand.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.ManaCost("{2}{U}"),
			Effect:  effects.Do(effects.Instruction(effects.StdB9ReturnThisToHand)),
		}},
	})
}
