package game

// cumulative_upkeep.go — #3869: the one fact another effect asks about
// cumulative upkeep (CR 702.24).
//
//	702.24a Cumulative upkeep is a triggered ability that imposes an
//	        increasing cost on a permanent.
//	702.24b If a permanent has multiple instances of cumulative upkeep,
//	        each triggers separately. However, the age counters are not
//	        connected to any particular ability; each cumulative upkeep
//	        ability will count the total number of age counters on the
//	        permanent at the time that ability resolves.
//
// Cumulative upkeep is not a canonicalKeywords token (it carries a
// cost, as ward does), so "has cumulative upkeep" cannot be read off
// Characteristic.Abilities. It is read off the triggers instead:
// effects.CumulativeUpkeepPaying stamps every instance with
// KeywordCumulativeUpkeep, and triggersOf returns the object's own
// triggers and the ones a layer-6 grant gave it (ADR 0093), so a
// printed instance and a granted one (Mana Chains, Balduvian Shaman)
// answer alike, and an object that lost its abilities (CR 613.1f)
// answers false.

// KeywordCumulativeUpkeep is the machine-readable name a cumulative
// upkeep trigger carries in TriggeredAbility.Keyword.
const KeywordCumulativeUpkeep = "cumulative upkeep"

// HasCumulativeUpkeep reports whether `c` has at least one instance of
// cumulative upkeep right now: printed, or granted by another effect.
// Balduvian Shaman's "that doesn't have cumulative upkeep" is its
// negation. Off the battlefield there is no layered characteristic, so
// a card there answers from what it prints.
func HasCumulativeUpkeep(c Card) bool { return CumulativeUpkeepInstances(c) > 0 }

// CumulativeUpkeepInstances counts the cumulative upkeep abilities `c`
// has (CR 702.24b: each instance triggers on its own).
func CumulativeUpkeepInstances(c Card) int {
	n := 0
	for _, t := range triggersOf(&c) {
		if t.Keyword == KeywordCumulativeUpkeep {
			n++
		}
	}
	return n
}
