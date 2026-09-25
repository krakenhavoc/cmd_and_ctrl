package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Astral Cornucopia — Artifact {X}{X}{X}:
//
//	"This artifact enters with X charge counters on it.
//	 {T}: Choose a color. Add one mana of that color for each charge
//	 counter on this artifact."
//
// Everflowing Chalice's shape (announced X onto a charge counter,
// XCounters) crossed with Gilded Lotus's colour pick (ProducedOneColor
// reading the counter count instead of a fixed 3). {X}{X}{X} still
// reads as one announced X — CastCounts.X is the single value, however
// many times the cost prints the symbol.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1bc42024-52da-4d93-8b47-544f0a4a72a1",
		Name:         "Astral Cornucopia",
		Completeness: CompletenessFull,
		XMatters:     true,
		EntersWithCountersFromCast: []game.EntryCountersFromCast{
			XCounters(game.CounterCharge),
		},
		ManaAbilities: []ManaAbility{{
			Cost:         ManaAbilityCost{Tap: true},
			ProducedFunc: ProducedOneColor(chargeCountersOn),
			Label:        "Choose a color. Add one mana of that color for each charge counter on this artifact",
		}},
	})
}

// chargeCountersOn reads the charge counters off the ability's own
// source, exactly as everflowingChaliceProduced does. Shared with The
// Eternity Elevator's 20+ mana ability, which reads the same counter
// off the same field.
func chargeCountersOn(g *game.Game, _, source uuid.UUID) int {
	c, ok := g.LookupCardForEffect(source)
	if !ok {
		return 0
	}
	return c.Counters[game.CounterCharge]
}
