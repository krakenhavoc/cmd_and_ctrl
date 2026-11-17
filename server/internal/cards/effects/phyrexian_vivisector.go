package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Phyrexian Vivisector — Creature — Phyrexian Human {1}{B}, 2/2 (Dominaria
// United):
//
//	"Whenever a creature you control dies, scry 1."
//
// Zulaport Cutthroat's trigger, which also fires for the Vivisector's own
// death (CR 603.10a: it sees itself leave). The scry is the footprinted
// primitive, and its fixed amount is declared on the card's Purpose.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9c99ce14-5059-47a9-b874-637b39473062",
		Name:         "Phyrexian Vivisector",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Scry: 1},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(WheneverACreatureYouControlDies("Phyrexian Vivisector — scry 1", Do(Scry{N: 1})), game.Purpose{DeathPayoff: true}),
		},
	})
}
