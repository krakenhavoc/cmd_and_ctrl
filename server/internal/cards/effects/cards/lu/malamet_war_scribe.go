package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Malamet War Scribe — Creature — Cat Warrior {3}{W}{W}, 4/3 (The Lost
// Caverns of Ixalan, #4470):
//
//	"When this creature enters, creatures you control get +2/+1 until end
//	 of turn."
//
// Invasion Tactics' enters pump with the printed numbers. The set of
// creatures is read once, as the trigger resolves (CR 611.2c), so it
// includes the Scribe itself and not a creature that enters later in the
// turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8978a30a-7918-4851-92ef-f40fdf90c379",
		Name:         "Malamet War Scribe",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Malamet War Scribe — creatures you control get +2/+1 until end of turn",
				effects.Do(effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Power: 2, Toughness: 1,
					Label: "Malamet War Scribe — creatures you control get +2/+1"})),
		},
	})
}
