package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b192_helpers.go — shared pieces of S82 GAP batch 192 (#5948): the
// Tyranid cards of Warhammer 40,000 Commander.

// RavenousEntry is the first half of ravenous (CR 702.156a): "this creature
// enters with X +1/+1 counters on it".
func RavenousEntry() []game.EntryCountersFromCast {
	return []game.EntryCountersFromCast{XCounters(game.CounterPlusOne)}
}

// RavenousDraw is the second half of ravenous (CR 702.156a): "if X is 5 or
// more, draw a card when it enters". An intervening-if enters trigger (CR
// 603.4) that reads the X announced for the spell (CR 107.3m), not the
// number of counters the creature actually got.
func RavenousDraw(name string) game.TriggeredAbility {
	return On(game.EventETB, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		return Self(ev, source, lki, g) && source.CastX() >= 5
	}, name+" — ravenous: X was 5 or more, draw a card", Do(DrawCards{N: 1}))
}

// TyranidWithCountersHaveHexproofAndIndestructible is the body of Tyrant
// Guard's Shieldwall: each creature its controller controls that has a counter
// on it gains hexproof and indestructible until end of turn. The set is fixed
// as the ability resolves (CR 611.2c).
func TyranidWithCountersHaveHexproofAndIndestructible(label string) Effect {
	return Do(GrantKeywordUntilEOT{
		Match: And(Creature(), YouControl(), func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
			return HasCounterOfKind(c, "")
		}),
		Keywords: []string{"hexproof", "indestructible"},
		Label:    label,
	})
}

// LosesAllAbilitiesAndHasIncludingLandMana is LosesAllAbilitiesAndHas for an
// effect that removes abilities without changing land types (Toxicrene's "all
// lands have '{T}: Add one mana of any color' and lose all other abilities"):
// the mana abilities a basic land type gives (CR 305.6) are abilities too, and
// go with the rest (CR 613.1f). The granted bundle is kept.
func LosesAllAbilitiesAndHasIncludingLandMana(applies func(target *game.Card, g *game.Game, source *game.Card) bool, keys ...string) game.StaticAbility {
	s := LosesAllAbilitiesAndHas(applies, keys...)
	s.Label = "Affected lands lose all abilities, mana abilities of their land types included, and have what the card says they have."
	apply := s.Apply
	s.Apply = func(c *game.Characteristic, target *game.Card, g *game.Game, source *game.Card) {
		if apply != nil {
			apply(c, target, g, source)
		}
		c.IntrinsicLandManaRemoved = true
	}
	return s
}
