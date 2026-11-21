package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nyxborn Hydra — Enchantment Creature — Hydra {X}{G}, 0/1 (Modern Horizons 3):
//
//	"Bestow {X}{G}{G}
//	 Reach, trample
//	 This permanent enters with X +1/+1 counters on it.
//	 Enchanted creature gets +1/+1 for each +1/+1 counter on this Aura and has
//	 reach and trample."
func init() {
	Register(Spec{
		OracleID:                   "4d1b655f-05b8-4b82-9cb9-67660f570a9b",
		Name:                       "Nyxborn Hydra",
		Completeness:               CompletenessFull,
		XMatters:                   true,
		PrintedKeywords:            []string{"reach", "trample"},
		AlternativeCosts:           []game.AlternativeCost{Bestow("{X}{G}{G}")},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{XCounters(game.CounterPlusOne)},
		Static: []game.StaticAbility{
			PumpAttachedPer(1, 1, func(_ *game.Game, source *game.Card) int { return source.Counters[game.CounterPlusOne] }),
			GrantToAttached("reach", "trample"),
		},
	})
}
