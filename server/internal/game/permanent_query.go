package game

import (
	"fmt"
	"slices"
	"strings"

	"github.com/google/uuid"
)

// permanent_query.go — ADR 0107's shared PermanentQuery: a closed, plain-
// data description of "a permanent with certain properties". §2 (#1879)
// is its first reader: "This creature can't attack unless defending player
// controls an Island" names the defending player's permanents with one.
//
// It is data and not a func because it lives inside a Characteristic
// (AttackTargetRestriction), and the snapshot's lastKnownBattlefield
// round-trips every Characteristic. A func there would be a closure route
// the ADR 0041 ratchet refuses.
//
// Every printed "defending player controls …" condition is one of these:
//
//	an Island, a Swamp, a Forest, a Mountain   Subtypes: {"Island"}
//	a snow land                                Supertypes: {"snow"}, Types: {"land"}
//	a blue permanent                           Colors: {"U"}
//	a creature with flying                     Types: {"creature"}, Keyword: "flying"
//	an enchantment or an enchanted permanent   two queries: {Types: {"enchantment"}}, {Enchanted: true}
//
// "Or" across whole queries is the caller's list (any of them); inside one
// query every set field must hold, and a list field is any of its entries.

// PermanentQuery is a closed filter over one permanent's EFFECTIVE
// characteristics. The zero value matches every permanent.
type PermanentQuery struct {
	// Types are lowercase card types ("land", "creature",
	// "enchantment"); the permanent must have at least one.
	Types []string `json:",omitempty"`
	// Subtypes are subtypes in their printed case ("Island"); the
	// permanent must have at least one.
	Subtypes []string `json:",omitempty"`
	// Supertypes are lowercase supertypes ("snow"); the permanent must
	// have every one.
	Supertypes []string `json:",omitempty"`
	// Colors are wire colours ("U"); the permanent must be at least one
	// of them.
	Colors []string `json:",omitempty"`
	// Keyword is a canonical keyword token ("flying") the permanent must
	// have.
	Keyword string `json:",omitempty"`
	// Enchanted asks for a permanent with an Aura attached to it.
	Enchanted bool `json:",omitempty"`
	// HasPowerAtMost asks for a permanent whose power is PowerAtMost or
	// less: "creatures with power 2 or less" (Queen Mother Ramonda,
	// #4659). A flag beside the number because zero is a real bound and a
	// ScopedEffect holds no pointers. Read live off the permanent's power
	// after the layers and its counters (CR 208.1, 613.4), not clamped, so
	// a -1/1 has power -1. Only a battlefield permanent has a power the
	// query can read: matchesCharacteristic never matches a bounded query.
	// Set the pair with WithPowerAtMost.
	HasPowerAtMost bool `json:",omitempty"`
	PowerAtMost    int  `json:",omitempty"`
	// HasPowerAtLeast is the other direction: "power 3 or greater".
	HasPowerAtLeast bool `json:",omitempty"`
	PowerAtLeast    int  `json:",omitempty"`
}

// WithPowerAtMost is q that also asks for power n or less.
func (q PermanentQuery) WithPowerAtMost(n int) PermanentQuery {
	q.HasPowerAtMost, q.PowerAtMost = true, n
	return q
}

// WithPowerAtLeast is q that also asks for power n or greater.
func (q PermanentQuery) WithPowerAtLeast(n int) PermanentQuery {
	q.HasPowerAtLeast, q.PowerAtLeast = true, n
	return q
}

// boundsPower reports whether q asks anything about power.
func (q PermanentQuery) boundsPower() bool { return q.HasPowerAtMost || q.HasPowerAtLeast }

// Equal reports whether q and o ask the same thing, field by field.
func (q PermanentQuery) Equal(o PermanentQuery) bool {
	return slices.Equal(q.Types, o.Types) &&
		slices.Equal(q.Subtypes, o.Subtypes) &&
		slices.Equal(q.Supertypes, o.Supertypes) &&
		slices.Equal(q.Colors, o.Colors) &&
		q.Keyword == o.Keyword &&
		q.Enchanted == o.Enchanted &&
		q.HasPowerAtMost == o.HasPowerAtMost && q.PowerAtMost == o.PowerAtMost &&
		q.HasPowerAtLeast == o.HasPowerAtLeast && q.PowerAtLeast == o.PowerAtLeast
}

// samePermanentQueries compares two any-of lists in order.
func samePermanentQueries(a, b []PermanentQuery) bool {
	return slices.EqualFunc(a, b, PermanentQuery.Equal)
}

// matchesLocked reports whether the battlefield permanent c matches q, by
// its effective characteristics: a land an effect made an Island is an
// Island (CR 305.7), and a granted flying is flying.
//
// Caller must hold g.mu with fresh layers.
func (q PermanentQuery) matchesLocked(g *Game, c *Card) bool {
	if c == nil {
		return false
	}
	if len(q.Types) > 0 && !slices.ContainsFunc(q.Types, c.HasCardType) {
		return false
	}
	if len(q.Subtypes) > 0 && !slices.ContainsFunc(q.Subtypes, c.HasSubtype) {
		return false
	}
	for _, s := range q.Supertypes {
		if !c.HasSupertype(s) {
			return false
		}
	}
	if len(q.Colors) > 0 && !slices.ContainsFunc(q.Colors, c.HasColor) {
		return false
	}
	if q.Keyword != "" && !HasKeyword(c, q.Keyword) {
		return false
	}
	if q.Enchanted && !g.isEnchantedLocked(c.InstanceID) {
		return false
	}
	if q.boundsPower() {
		p := c.PowerForComparison()
		if (q.HasPowerAtMost && p > q.PowerAtMost) || (q.HasPowerAtLeast && p < q.PowerAtLeast) {
			return false
		}
	}
	return true
}

// isEnchantedLocked reports whether an Aura is attached to the permanent
// `id`. An Equipment does not enchant.
//
// Caller must hold g.mu.
func (g *Game) isEnchantedLocked(id uuid.UUID) bool {
	for i := range g.Battlefield.Cards {
		a := &g.Battlefield.Cards[i]
		if a.IsAttachedTo(id) && a.IsAura() {
			return true
		}
	}
	return false
}

// controlsMatchingLocked reports whether `player` controls a permanent on
// the battlefield that matches any of `qs`. Phased-out permanents are not
// on the battlefield (CR 702.26b), so they never count.
//
// Caller must hold g.mu with fresh layers.
func (g *Game) controlsMatchingLocked(player uuid.UUID, qs []PermanentQuery) bool {
	if player == uuid.Nil {
		return false
	}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller != player {
			continue
		}
		for _, q := range qs {
			if q.matchesLocked(g, c) {
				return true
			}
		}
	}
	return false
}

// countMatchingLocked counts the permanents `player` controls on the
// battlefield that match any of `qs`, each once: "you control more
// creatures than defending player".
//
// Caller must hold g.mu with fresh layers.
func (g *Game) countMatchingLocked(player uuid.UUID, qs []PermanentQuery) int {
	if player == uuid.Nil {
		return 0
	}
	n := 0
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller != player {
			continue
		}
		for _, q := range qs {
			if q.matchesLocked(g, c) {
				n++
				break
			}
		}
	}
	return n
}

// CountMatchingForEffect is countMatchingLocked for card code: the
// permanents `player` controls that match any of `qs`. Read-only.
//
// Caller must hold g.mu with fresh layers (a block rule's Pair does).
func (g *Game) CountMatchingForEffect(player uuid.UUID, qs []PermanentQuery) int {
	return g.countMatchingLocked(player, qs)
}

// permanentQueriesPlural is PermanentQueriesNoun in the plural, for
// "more creatures than Bob": each query's head noun takes the "s".
func permanentQueriesPlural(qs []PermanentQuery) string {
	parts := make([]string, 0, len(qs))
	for _, q := range qs {
		n := q.Noun()
		if head, tail, ok := strings.Cut(n, " with "); ok {
			parts = append(parts, head+"s with "+tail)
		} else {
			parts = append(parts, n+"s")
		}
	}
	return strings.Join(parts, " or ")
}

// Noun is the query as a player reads it, without an article: "Island",
// "snow land", "blue permanent", "creature with flying", "enchanted
// permanent". It is the chip's and the refusal's wording ("Bob controls no
// Island").
func (q PermanentQuery) Noun() string {
	var words []string
	if q.Enchanted {
		words = append(words, "enchanted")
	}
	for _, c := range q.Colors {
		words = append(words, ColorName(c))
	}
	words = append(words, q.Supertypes...)
	switch {
	case len(q.Subtypes) > 0:
		words = append(words, strings.Join(q.Subtypes, " or "))
	case len(q.Types) > 0:
		words = append(words, strings.Join(q.Types, " or "))
	default:
		words = append(words, "permanent")
	}
	s := strings.Join(words, " ")
	var with []string
	if q.Keyword != "" {
		with = append(with, q.Keyword)
	}
	switch {
	case q.HasPowerAtMost && q.HasPowerAtLeast:
		with = append(with, fmt.Sprintf("power %d to %d", q.PowerAtLeast, q.PowerAtMost))
	case q.HasPowerAtMost:
		with = append(with, fmt.Sprintf("power %d or less", q.PowerAtMost))
	case q.HasPowerAtLeast:
		with = append(with, fmt.Sprintf("power %d or greater", q.PowerAtLeast))
	}
	if len(with) > 0 {
		s += " with " + strings.Join(with, " and ")
	}
	return s
}

// PermanentQueriesNoun joins an any-of list the way a card prints it:
// "enchantment or enchanted permanent".
func PermanentQueriesNoun(qs []PermanentQuery) string {
	parts := make([]string, 0, len(qs))
	for _, q := range qs {
		parts = append(parts, q.Noun())
	}
	return strings.Join(parts, " or ")
}
