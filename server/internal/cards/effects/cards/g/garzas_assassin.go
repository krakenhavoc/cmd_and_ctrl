package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Garza's Assassin — Creature — Human Assassin {B}{B}{B}, 2/2 (Coldsnap):
//
//	"Sacrifice this creature: Destroy target nonblack creature.
//	 Recover—Pay half your life, rounded up."
//
// The sacrifice is a cost (CR 118.3), so it needs no tap and works the turn
// the Assassin arrives. Recover (CR 702.59a) is S82OldB564RecoverHalfLife:
// from the graveyard, when a creature is put into its owner's graveyard from
// the battlefield, the owner may pay half their life, rounded up, to return it
// to hand; otherwise it is exiled. The Assassin itself, entering the graveyard
// as it is sacrificed, never triggers its own recover.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9660b7c4-7b04-4ec3-b867-299c141f77e6",
		Name:         "Garza's Assassin",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this creature: Destroy target nonblack creature.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.SacrificeThis(),
			Targets: effects.TargetCreature("target nonblack creature", effects.NonBlack()),
			Effect:  effects.DestroyFirstLegalTarget,
		}},
		Triggered: []game.TriggeredAbility{effects.S82OldB564RecoverHalfLife("Garza's Assassin")},
	})
}
