package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fated Firepower — Enchantment {X}{R}{R}{R} (EDHREC rank 2985):
//
//	"Flash
//	 This enchantment enters with X fire counters on it.
//	 If a source you control would deal damage to an opponent or a
//	 permanent an opponent controls, it deals that much damage plus
//	 an amount of damage equal to the number of fire counters on this
//	 enchantment instead."
//
// The Duskmourn damage amplifier. Flash rides PrintedKeywords (the
// cast path reads it off a card in hand). The fire counters are the
// printed CR 614.1c entry clause and ride the CR 614 pipeline as one
// (XCounters, #1002), and the replacement adds the count read AT THE
// MOMENT THE DAMAGE WOULD BE DEALT — a proliferate that lands after
// the Firepower grows every later hit — to any damage from a source
// the controller controls
// aimed at an opponent or one of their permanents: combat damage, a
// Bolt, a pinger, a fight (b28AddCountersToDamageAtOpponents). Damage
// to the controller's own permanents, or to themselves, is untouched,
// as printed. With zero fire counters the replacement does not apply
// at all. CR 616: alongside a doubler the affected player orders the
// two, so "+N then ×2" and "×2 then +N" are both reachable.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:                   "13daa21c-278d-45bd-9a6e-a77d6a558453",
		Name:                       "Fated Firepower",
		Completeness:               CompletenessFull,
		PrintedKeywords:            []string{"flash"},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{XCounters("fire")},
		Replacements: []game.ReplacementEffect{
			b28AddCountersToDamageAtOpponents("Fated Firepower: extra damage for each fire counter", "fire"),
		},
	})
}
