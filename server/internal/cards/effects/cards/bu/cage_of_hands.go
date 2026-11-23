package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cage of Hands — {2}{W} Enchantment — Aura: "Enchant creature.
// Enchanted creature can't attack or block. {1}{W}: Return this Aura to
// its owner's hand."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "18ffd8ad-1556-450d-8bc1-8103a3d0404e",
		Name:         "Cage of Hands",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			effects.RestrictAttached(game.CantAttackOrBlock),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{W}: Return this Aura to its owner's hand.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.ManaCost("{1}{W}"),
			Effect:  effects.Do(effects.Instruction(effects.StdB9ReturnThisToHand)),
		}},
	})
}
