package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wall of Nets — Creature — Wall {1}{W}{W}, 0/7 (Exodus, S82 OLD batch 744, #5626):
//
//	"Defender
//	 At end of combat, exile all creatures blocked by this creature.
//	 When this creature leaves the battlefield, return all cards exiled with it
//	 to the battlefield under their owners' control."
//
// Worldgorger Dragon's linked pair. At the beginning of each combat's end, every
// creature the Wall is still blocking is exiled in one move, the exile recorded
// under the trigger's label (CR 607.2a), so it accumulates over several combats;
// and when the Wall leaves, whatever that record still holds in exile comes back
// together under its owners' control (ReturnFromExileTogether). Tokens cease to
// exist in exile and do not return.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "cd112962-b297-403c-bede-c8b587f379b3",
		Name:            "Wall of Nets",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Triggered: []game.TriggeredAbility{
			effects.AtEachStep(game.StepEndCombat, wallOfNetsExileLabel, effects.S82OldB744ExileBlockedCreatures),
			effects.On(game.EventLTB, effects.Self, "Wall of Nets — return all cards exiled with it to the battlefield under their owners' control",
				effects.B41ReturnCardsExiledWithToTheBattlefield(wallOfNetsExileLabel)),
		},
	})
}

const wallOfNetsExileLabel = "Wall of Nets — exile all creatures blocked by this creature"
