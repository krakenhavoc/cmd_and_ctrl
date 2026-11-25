package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Phoenix of Ash — Creature — Phoenix {1}{R}{R}, 2/2:
//
//	"Flying, haste
//	 {2}{R}: Phoenix of Ash gets +2/+0 until end of turn.
//	 Escape—{2}{R}{R}, Exile three other cards from your graveyard.
//	 Phoenix of Ash escapes with a +1/+1 counter on it."
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "0ad557fc-6b7e-4f7e-b060-e3c8ec8acf6d",
		Name:             "Phoenix of Ash",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"flying", "haste"},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.EscapeWithCounters("{2}{R}{R}", 3, 1)},
		Activated:        []effects.ActivatedAbility{effects.ThisGetsRow("{2}{R}", 2, 0, "Phoenix of Ash")},
	})
}
