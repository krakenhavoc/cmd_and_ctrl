package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Soul Snare — Enchantment {W}:
//
//	"{W}, Sacrifice this enchantment: Exile target creature that's
//	 attacking you or a planeswalker you control."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "52bd4331-67b6-4eb8-9566-e00359889fc6",
		Name:         "Soul Snare",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}, Sacrifice this enchantment: Exile target creature that's attacking you or a planeswalker you control.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.Plus(effects.ManaCost("{W}"), effects.SacrificeThis()),
			Targets: effects.TargetCreature("target creature that's attacking you or a planeswalker you control", effects.Cmdr2B03AttackingYouOrYourPlaneswalker()),
			Effect:  effects.ExileFirstTarget,
		}},
	})
}
