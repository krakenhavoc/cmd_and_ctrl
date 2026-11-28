package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hunting Kavu (Invasion, S82 OLD batch 646, #5616):
//
//	Creature — Kavu {1}{R}{G}, 2/3:
//	"{1}{R}{G}, {T}: Exile this creature and target creature without flying that's attacking you."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a1b7fbfc-c510-4b1c-a2b9-c1204c699393",
		Name:         "Hunting Kavu",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}{G}, {T}: Exile this creature and target creature without flying that's attacking you.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.Plus(effects.ManaCost("{1}{R}{G}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature without flying that's attacking you", effects.AttackingYou(), effects.WithoutKeyword("flying")),
			Effect:  effects.S82OldB646ExileThisAndTarget,
		}},
	})
}
