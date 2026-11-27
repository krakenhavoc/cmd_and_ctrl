package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mine Bearer — Creature — Human Soldier {2}{W}, 1/1 (Prophecy, S82 OLD b672, #5617):
//
//	"{T}, Sacrifice this creature: Destroy target attacking creature."
//
// Barbarian Riftcutter's sacrifice-as-a-cost destroy with Kill Shot's target: the
// creature is paid as the cost, so the ability resolves with its source already
// gone, and a target that stopped attacking (or left) in response is not
// destroyed (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "19fba3c8-3363-4f2f-a2fb-01e59a77447c",
		Name:         "Mine Bearer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice this creature: Destroy target attacking creature.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.Plus(effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetCreature("target attacking creature", effects.AttackingCreature()),
			Effect:  effects.DestroyChosenPermanent,
		}},
	})
}
