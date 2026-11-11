package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Soul Snare — Enchantment {W}:
//
//	"{W}, Sacrifice this enchantment: Exile target creature that's
//	 attacking you or a planeswalker you control."
func init() {
	Register(Spec{
		OracleID:     "52bd4331-67b6-4eb8-9566-e00359889fc6",
		Name:         "Soul Snare",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{W}, Sacrifice this enchantment: Exile target creature that's attacking you or a planeswalker you control.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    Plus(ManaCost("{W}"), SacrificeThis()),
			Targets: TargetCreature("target creature that's attacking you or a planeswalker you control", cmdr2B03AttackingYouOrYourPlaneswalker()),
			Effect:  ExileFirstTarget,
		}},
	})
}
