package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cage of Hands — {2}{W} Enchantment — Aura: "Enchant creature.
// Enchanted creature can't attack or block. {1}{W}: Return this Aura to
// its owner's hand."
func init() {
	Register(Spec{
		OracleID:     "18ffd8ad-1556-450d-8bc1-8103a3d0404e",
		Name:         "Cage of Hands",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			RestrictAttached(game.CantAttackOrBlock),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{W}: Return this Aura to its owner's hand.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    ManaCost("{1}{W}"),
			Effect:  Do(Instruction(stdB9ReturnThisToHand)),
		}},
	})
}
