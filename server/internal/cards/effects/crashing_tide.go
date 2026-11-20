package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Crashing Tide — Sorcery {2}{U} (Rivals of Ixalan, #5531):
//
//	"This spell has flash as long as you control a Merfolk.
//	 Return target creature to its owner's hand.
//	 Draw a card."
//
// Repulse's body at sorcery speed, with Supernatural Rescue's flash
// permission (Spec.FlashIf, CR 601.3) read as the cast begins: while you
// control a Merfolk it may be cast any time you could cast an instant.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4c3f87d7-a632-4b8c-9052-b4055ea88771",
		Name:         "Crashing Tide",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		FlashIf: FlashIfYou("as long as you control a Merfolk",
			func(g *game.Game, you uuid.UUID) bool { return b05ControlsSubtype(g, you, "Merfolk") }),
		Targets:   TargetCreature("target creature"),
		OnResolve: bounceTargetThenDrawACard,
	})
}
