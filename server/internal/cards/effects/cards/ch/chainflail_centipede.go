package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chainflail Centipede — Artifact Creature — Equipment Insect {2}{B}, 2/2:
//
//	"Whenever this creature or equipped creature attacks, it gets +2/+0
//	 until end of turn.
//	 Reconfigure {2}"
//
// "It" is the creature that attacked: the Centipede while it fights on
// its own, its host while it is attached (an attached reconfigure
// Equipment is not a creature, CR 702.151b, so it never attacks). An
// attacker that left before the trigger resolved gets nothing.
// Reconfigure is reconfigure.go (#2639).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e77a1809-d21a-4bb6-88a6-35eb13771b1d",
		Name:         "Chainflail Centipede",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, effects.ThisOrEquippedCreatureAttacks,
				"Chainflail Centipede — it gets +2/+0 until end of turn", chainflailCentipedePump),
		},
		Activated: effects.Reconfigure("{2}"),
	})
}

func chainflailCentipedePump(g *game.Game, item *game.StackItem) error {
	return effects.PumpTriggeringCreatureUntilEOT(effects.NewContext(g, item), 2, 0, "Chainflail Centipede — +2/+0")
}
