package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Room(effects.RoomSpec{
		OracleID:     "2b46394e-d337-4b1a-88e6-7fdac2ee4ac4",
		Name:         "Underwater Tunnel // Slimy Aquarium",
		Completeness: effects.CompletenessFull,
		Left: effects.Door{Triggered: []game.TriggeredAbility{
			effects.WhenYouUnlockThisDoor(game.DoorLeft, "Underwater Tunnel — surveil 2", effects.Do(effects.Surveil{N: 2})),
		}},
		Right: effects.Door{Triggered: []game.TriggeredAbility{
			effects.WhenYouUnlockThisDoor(game.DoorRight, "Slimy Aquarium — manifest dread, then put a +1/+1 counter on that creature",
				effects.Do(effects.ManifestDread{ThenTail: effects.PutCountersOnManifested(effects.CounterAmount{Kind: game.CounterPlusOne, N: 1})})),
		}},
	}))
}
