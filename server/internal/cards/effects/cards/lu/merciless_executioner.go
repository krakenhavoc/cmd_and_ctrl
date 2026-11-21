package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Merciless Executioner — Creature — Orc Warrior {2}{B}, 3/1 (EDHREC
// rank 2205):
//
//	"When this creature enters, each player sacrifices a creature of
//	 their choice."
//
// Fleshbag Marauder with a different type line, and the same body:
// EachPlayerSacrifices with the controller included — "each player"
// means you too, and the Executioner is on the battlefield when its
// own trigger resolves, so feeding it to itself is the printed line.
// Every player picks their own creature in their own prompt; a
// player with none is skipped (CR 701.21a).
//
// The ETB is a triggered ability and uses the stack (#578). It used to
// run from the direct AsEnters hook, which gave nobody a response
// window; now the trigger waits for every player to pass, like every
// other "When ~ enters".
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c3c45d50-9038-41df-bb2f-9bc40071845b",
		Name:         "Merciless Executioner",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerEdict(effects.WhenThisEnters("Merciless Executioner — each player sacrifices a creature", effects.Do(effects.EachPlayerSacrifices{
				Match: effects.Creature(),
				Label: "a creature",
			}))),
		},
	})
}
