package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Underwater Tunnel // Slimy Aquarium — Enchantment — Room (ADR 0103):
//
//	Underwater Tunnel {U}: "When you unlock this door, surveil 2."
//	Slimy Aquarium {3}{U}: "When you unlock this door, manifest dread,
//	 then put a +1/+1 counter on that creature."
//
// Underwater Tunnel is the Surveil primitive. Slimy Aquarium is
// manifest dread (CR 701.62a, ADR 0082's 2026-10-07 amendment) with the
// counter as its continuation, on the creature that entered.
//
// No simplification.
func init() {
	Register(Room(RoomSpec{
		OracleID:     "2b46394e-d337-4b1a-88e6-7fdac2ee4ac4",
		Name:         "Underwater Tunnel // Slimy Aquarium",
		Completeness: CompletenessFull,
		Left: Door{Triggered: []game.TriggeredAbility{
			WhenYouUnlockThisDoor(game.DoorLeft, "Underwater Tunnel — surveil 2", Do(Surveil{N: 2})),
		}},
		Right: Door{Triggered: []game.TriggeredAbility{
			WhenYouUnlockThisDoor(game.DoorRight, "Slimy Aquarium — manifest dread, then put a +1/+1 counter on that creature",
				Do(ManifestDread{Then: PutCountersOnManifested(CounterAmount{Kind: game.CounterPlusOne, N: 1})})),
		}},
	}))
}
