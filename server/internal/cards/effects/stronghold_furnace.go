package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Stronghold Furnace — Plane — Rath (Commander Masters, #4870):
//
//	"If a source would deal damage to a permanent or player, it deals
//	 double that damage instead.
//	 Whenever chaos ensues, Stronghold Furnace deals 1 damage to any
//	 target."
//
// A plane (ADR 0216 §6). The replacement doubles every source's damage,
// whoever's it is (Dictate of the Twin Gods' clause, from the command
// zone: CR 901.7); the plane's own chaos damage is doubled by it, so the
// 1 damage is 2. The chaos ability (CR 311.7) is dealt by the plane.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ff76b576-8ad3-44f9-8043-80003a078e84",
		Name:         "Stronghold Furnace",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{allDamageDoubled("Stronghold Furnace: double damage")},
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverChaosEnsues("Stronghold Furnace — deals 1 damage to any target", DamageToFirstTarget(1)), TargetAny()),
		},
	})
}

// allDamageDoubled is "If a source would deal damage to a permanent or
// player, it deals double that damage instead": every source, everyone's
// damage.
func allDamageDoubled(label string) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventDealDamage},
		AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) bool {
			return ev.Kind == game.RepEventDamage && ev.DamageAmount > 0
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.DamageAmount *= 2
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label: label,
	}
}
