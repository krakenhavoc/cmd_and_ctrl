package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Abyssal Gorestalker — Creature — Horror {4}{B}{B}, 6/6 (The Lost
// Caverns of Ixalan, #4470):
//
//	"When this creature enters, each player sacrifices two creatures of
//	 their choice."
//
// Barter in Blood on an enters trigger: each player, the Gorestalker's
// controller included, chooses two creatures in APNAP order (CR 101.4)
// and the pair is sacrificed together (CR 701.21a). The Gorestalker is
// on the battlefield when its own trigger resolves, so it is a legal
// choice for its controller. A player with one creature sacrifices it; a
// player with none is not asked. Nothing is targeted.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "465cda20-c8f3-4a7c-a250-875877bce21d",
		Name:         "Abyssal Gorestalker",
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepSacrifice, Partial: true}},
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerEdict(WhenThisEnters("Abyssal Gorestalker — each player sacrifices two creatures",
				jmpB1EachPlayerSacrificesN(2, "Abyssal Gorestalker — choose two creatures to sacrifice"))),
		},
	})
}
