package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// GrantedAltCostZeroUpToTimeCounters is As Foretold's "Once each turn, you may
// pay {0} rather than pay the mana cost for a spell you cast with mana value X
// or less, where X is the number of time counters on this enchantment" (#5929).
// Its own key: a claim is stored by key. An on-disk identity, never renamed.
const GrantedAltCostZeroUpToTimeCounters = "granted-zero-up-to-time-counters"

// PayZeroForASpellUpToTimeCountersOnceEachTurn is As Foretold's static. Free,
// so X is 0 (CR 107.3b); any zone the printed mana cost could be paid from
// (CR 118.9a); the cap is the number of time counters on the granting
// permanent as the offer is derived (MaxManaValueOfSource), and the limit is
// per source, spent as the spell becomes cast. The mana value is the spell's
// as the offer replaces its mana cost, an {X} counting as 0 (CR 107.3b).
func PayZeroForASpellUpToTimeCountersOnceEachTurn() game.GrantedAlternativeCost {
	return game.GrantedAlternativeCost{
		Offer: game.AlternativeCost{
			Key:   GrantedAltCostZeroUpToTimeCounters,
			Label: "Pay {0} rather than pay this spell's mana cost",
		},
		OnceEachTurn: true,
		MaxManaValueOfSource: func(_ *game.Game, source *game.Card) int {
			return source.Counters[game.CounterTime]
		},
		MaxManaValueText: "the number of time counters on this enchantment",
	}
}
