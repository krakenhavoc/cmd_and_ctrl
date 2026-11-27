package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Celestial Kirin — Legendary Creature — Kirin Spirit {2}{W}{W}, 3/3 (Saviors of
// Kamigawa, S82 OLD b593):
//
//	"Flying
//	 Whenever you cast a Spirit or Arcane spell, destroy all permanents with that
//	 spell's mana value."
//
// Fires on casting, once per spell. The number is the spell's mana value as it was cast
// (X counted, CR 202.3e), read from the cast event's snapshot (CR 603.10, 608.2h), so a
// spell countered in response still sets it. Every permanent with that mana value is
// destroyed, the Kirin's own and a token's (mana value 0) included; lands are mana value
// 0 as well.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a5f7967e-d357-4ab6-96ac-77262864f280",
		Name:            "Celestial Kirin",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{Sweep: game.Sweep{Matches: game.SweepAllPermanents, How: game.SweepDestroy, Partial: true}},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.ChkB1SpiritOrArcane(), "Celestial Kirin — destroy all permanents with that spell's mana value", effects.CelestialKirinDestroy),
		},
	})
}
