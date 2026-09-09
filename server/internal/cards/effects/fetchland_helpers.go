package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// fetchland_helpers.go — the shared body of the sacrifice-to-fetch
// land family, from the frequency-ordered Commander staples pass.
//
// Lives in its own file rather than in helpers.go so a concurrent
// card batch editing that file doesn't collide with this one — the
// convention aang_helpers.go and pirates_batch2_helpers.go set.
//
// One sandbox gap runs through every card here, and it is worth
// stating once rather than ten times: SearchLibrary takes the FIRST
// match in library order, and choosing WHICH dual to fetch is a
// fetchland's entire strategic content. Rampant Growth, Cultivate,
// Nature's Lore and Three Visits already ship with that
// simplification, so nothing new is conceded — but it bites harder
// here, and a search chooser is now the highest-value gap standing
// between this catalog and the top of the play-rate list.

// fetchlandCost is "{T}, Pay 1 life, Sacrifice this land" — the
// activation cost shared by the ten Zendikar / Onslaught
// fetchlands.
//
// No SorcerySpeed: cracking a fetch on an opponent's end step is
// the normal line, not an edge case. A land has no summoning
// sickness, so the Tap component is payable the turn it is played,
// which is also correct.
func fetchlandCost() game.AbilityCost {
	return Plus(TapCost(), PayLife(1), SacrificeThis())
}

// landWithEitherSubtype matches any land carrying either subtype —
// "an Island or Swamp card". It composes IsLandWithSubtype rather
// than re-implementing the scan so the two share one matching
// posture, and it deliberately admits NONBASIC lands: in paper a
// Polluted Delta fetches Watery Grave far more often than it
// fetches a basic Island, and the printed text says "card", not
// "basic land card".
//
// Needles MUST be lowercase. containsFoldASCII folds the haystack
// and not the needle, so "Island" would silently match nothing —
// the trap already documented on discardedCardHasType.
func landWithEitherSubtype(a, b string) func(game.Card) bool {
	first, second := IsLandWithSubtype(a), IsLandWithSubtype(b)
	return func(c game.Card) bool { return first(c) || second(c) }
}

// fetchDual is the effect body of a Zendikar / Onslaught fetchland.
// The land arrives untapped, which is the whole reason these cost a
// life and Evolving Wilds does not.
func fetchDual(a, b string) func(*game.Game, *game.StackItem) error {
	pred := landWithEitherSubtype(a, b)
	return func(g *game.Game, item *game.StackItem) error {
		return SearchLibrary{
			Player:    item.Controller,
			Predicate: pred,
			Dest:      game.ZoneBattlefield,
			Limit:     1,
			Reveal:    true,
			Shuffle:   true,
		}.Apply(NewContext(g, item))
	}
}

// fetchBasicTapped is the effect body of Evolving Wilds and
// Terramorphic Expanse: any basic land, onto the battlefield
// tapped. The two cards are functionally identical and differ only
// in name and oracle ID.
func fetchBasicTapped(g *game.Game, item *game.StackItem) error {
	return SearchLibrary{
		Player:        item.Controller,
		Predicate:     IsBasicLand,
		Dest:          game.ZoneBattlefield,
		Limit:         1,
		Reveal:        true,
		Shuffle:       true,
		TappedOnEntry: true,
	}.Apply(NewContext(g, item))
}
