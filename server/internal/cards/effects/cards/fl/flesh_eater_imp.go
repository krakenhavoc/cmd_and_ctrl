package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flesh-Eater Imp — Creature — Phyrexian Imp {3}{B}, 2/2 (Mirrodin Besieged,
// #5576):
//
//	"Flying
//	 Infect (This creature deals damage to creatures in the form of
//	 -1/-1 counters and to players in the form of poison counters.)
//	 Sacrifice a creature: This creature gets +1/+1 until end of turn."
//
// Devouring Swarm's sacrifice-a-creature self pump on an infect flyer. The
// Imp may sacrifice itself, and then the pump has nothing to land on.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b1aae085-2dca-425c-911a-19a89a0a4993",
		Name:            "Flesh-Eater Imp",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "infect"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice a creature: This creature gets +1/+1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerSacOutlet},
			Cost:    effects.SacrificeACreature(),
			Effect:  effects.ThisGetsUntilEndOfTurn(1, 1, "Flesh-Eater Imp — +1/+1 until end of turn"),
		}},
	})
}
