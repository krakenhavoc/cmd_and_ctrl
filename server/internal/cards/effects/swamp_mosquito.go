package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Swamp Mosquito — Creature — Insect {1}{B}, 0/1 (EDHREC rank 19346):
//
//	"Flying
//	 Whenever this creature attacks and isn't blocked, defending
//	 player gets a poison counter."
//
// A proof card for #1279. "Attacks and isn't blocked" (CR 509.3) fires
// off the defending player's COMPLETED block declaration
// (EventBlockersDeclared, WhenAttacksAndIsNotBlocked); before the
// declaration had a completion point there was no moment to fire on,
// because "not blocked" and "not asked yet" were the same empty
// record. The counter is placed by the trigger's controller through
// the CR 614 player-counter window, as Ichor Rats places its own.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:        "4bb4844c-2678-45c7-8f5e-9cf185fd484b",
		Name:            "Swamp Mosquito",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenAttacksAndIsNotBlocked("Swamp Mosquito — defending player gets a poison counter", defendingPlayerGetsAPoisonCounter),
		},
	})
}

// defendingPlayerGetsAPoisonCounter gives the defending player one
// poison counter, placed by the trigger's controller.
func defendingPlayerGetsAPoisonCounter(g *game.Game, item *game.StackItem, defender uuid.UUID) error {
	if !defendingPlayerStillIn(g, defender) {
		return nil
	}
	return g.AddPlayerCounterByForEffect(item.Controller, defender, game.CounterPoison, 1)
}
