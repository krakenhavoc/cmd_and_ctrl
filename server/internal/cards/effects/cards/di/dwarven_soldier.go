package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dwarven Soldier — Creature — Dwarf Soldier {1}{R}, 2/1 (Fallen Empires, S82
// OLD b854, #5637):
//
//	"Whenever this creature blocks or becomes blocked by one or more Orcs, this
//	 creature gets +0/+2 until end of turn."
//
// One trigger per block, not per pair (CR 509.3a, 509.3c), as Amphibious Kavu's:
// see S82OldB854BlocksOrIsBlockedBySubtype.
//
// No simplification.
func init() {
	const label = "Dwarven Soldier — this creature gets +0/+2 until end of turn"
	effects.Register(effects.Spec{
		OracleID:     "88f9075c-ea96-4f8d-b0d1-e5e244e824c1",
		Name:         "Dwarven Soldier",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventBlock, game.EventBecomesBlocked}, effects.S82OldB854BlocksOrIsBlockedBySubtype("Orc"),
				label, effects.ThisGetsUntilEndOfTurn(0, 2, label)),
		},
	})
}
