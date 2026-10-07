package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Underwater Tunnel // Slimy Aquarium — Enchantment — Room (ADR 0103):
//
//	Underwater Tunnel {U}: "When you unlock this door, surveil 2."
//	Slimy Aquarium {3}{U}: "When you unlock this door, manifest dread,
//	 then put a +1/+1 counter on that creature."
//
// Underwater Tunnel is the Surveil primitive. Slimy Aquarium is NOT
// implemented: manifest dread (CR 701.62a) has no keyword action in the
// engine, so the door is empty. That is weaker than printed, and the
// card says so.
func init() {
	Register(Room(RoomSpec{
		OracleID:     "2b46394e-d337-4b1a-88e6-7fdac2ee4ac4",
		Name:         "Underwater Tunnel // Slimy Aquarium",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Slimy Aquarium's manifest dread isn't implemented, so unlocking it does nothing."},
		Left: Door{Triggered: []game.TriggeredAbility{
			WhenYouUnlockThisDoor(game.DoorLeft, "Underwater Tunnel — surveil 2", Do(Surveil{N: 2})),
		}},
	}))
}
