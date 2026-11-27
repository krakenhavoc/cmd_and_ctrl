package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Empty-Shrine Kannushi — Creature — Human Cleric {W}, 1/1 (Betrayers of Kamigawa, S82 OLD
// b595):
//
//	"This creature has protection from each color among permanents you control."
//
// A layer-6 static on itself that adds one protection quality per colour found among the
// permanents its controller controls (CR 702.16), the Kannushi's own white included;
// Etched Champion's shape with a computed set. It follows the board as colours come and go.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "39a32629-8b05-4539-9ce7-1a6e467d0064",
		Name:         "Empty-Shrine Kannushi",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.OldB595ProtectionFromColorsAmongYourPermanents()},
	})
}
