package game

import (
	"sort"

	"github.com/google/uuid"
)

// keyword_counters.go — CR 122.1b, ADR 0101 (#1753): "A keyword
// counter on a permanent or on a card in a zone other than the
// battlefield causes that object to gain that keyword."
//
// A keyword counter is a Card.Counters entry whose kind IS a canonical
// keyword token ("flying", "first strike", "indestructible"). The
// engine reads it in three places and nowhere else:
//
//   - the layer pass, where each (permanent, keyword kind) is one
//     layer-6 continuous effect at the counter's own CR 613.7c
//     timestamp (keywordCounterEffectsLocked, below). CR 613.1f puts
//     keyword counters in layer 6 beside grants and removals, and
//     CR 613.3 orders them all by timestamp, so a "loses all
//     abilities" older than the counter leaves the keyword and one
//     newer than it takes the keyword away. The effect has no source
//     permanent, so CR 613.6's silencing never touches it: a counter
//     is not an ability of the permanent it sits on.
//   - the off-battlefield ability walk (forEachAbilityToken), for a
//     card in another zone that an effect put a keyword counter on.
//   - applyCounterByLocked, the one counter door, which stamps the
//     timestamp (stampKeywordCounter).
//
// Everything downstream — HasKeyword, targeting, combat, destruction,
// the legal-move enumerator, the bot and the wire — reads the one
// effective ability list and needed no change (ADR 0046).

// Card-level keyword counter kinds (CR 122.1b). Each is the canonical
// keyword token it grants, spelled exactly as canonicalKeywords spells
// it, so a card file writes game.CounterFlying and the engine needs no
// name mapping.
const (
	CounterFlying         = "flying"
	CounterFirstStrike    = "first strike"
	CounterDoubleStrike   = "double strike"
	CounterDeathtouch     = "deathtouch"
	CounterExalted        = "exalted"
	CounterHaste          = "haste"
	CounterHexproof       = "hexproof"
	CounterIndestructible = "indestructible"
	CounterLifelink       = "lifelink"
	CounterMenace         = "menace"
	CounterReach          = "reach"
	CounterShadow         = "shadow"
	CounterTrample        = "trample"
	CounterVigilance      = "vigilance"
)

// keywordCounterKinds is CR 122.1b's list INTERSECTED with
// canonicalKeywords. The table is closed, like the keyword table it is
// drawn from: a kind joins it in the same change that teaches the
// engine to honour its keyword. One of CR 122.1b's fifteen is out for
// that reason — decayed is not a canonical keyword yet (ADR 0101 owner
// decision 5, #2650) — and so is every "hexproof from [quality]"
// variant, which ADR 0038 §6 refused as a keyword. Exalted joined with
// #2538 (ADR 0101 amendment 2026-10-08).
//
// TestKeywordCounterKindsAreCR1221b and
// TestKeywordCounterKindsAreCanonicalKeywords hold it to both halves.
var keywordCounterKinds = map[string]bool{
	CounterFlying:         true,
	CounterFirstStrike:    true,
	CounterDoubleStrike:   true,
	CounterDeathtouch:     true,
	CounterExalted:        true,
	CounterHaste:          true,
	CounterHexproof:       true,
	CounterIndestructible: true,
	CounterLifelink:       true,
	CounterMenace:         true,
	CounterReach:          true,
	CounterShadow:         true,
	CounterTrample:        true,
	CounterVigilance:      true,
}

// IsKeywordCounter reports whether a counter kind is a CR 122.1b
// keyword counter the engine reads. It is the ONE predicate for the
// question: the stamp, the layer pass and the off-battlefield walk all
// ask it. A kind outside the table — a misspelling, "Flying", a
// homebrew counter — is storage and grants nothing.
func IsKeywordCounter(kind string) bool {
	return keywordCounterKinds[kind]
}

// KeywordCounterKinds returns the table's kinds, sorted. A copy: the
// table stays closed.
func KeywordCounterKinds() []string {
	out := make([]string, 0, len(keywordCounterKinds))
	for k := range keywordCounterKinds {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// stampKeywordCounter keeps Card.CounterStampedAt in step with a
// counter change of `delta` that left `count` counters of `kind` on
// the card. CR 613.7c: a placement gives EVERY counter of the kind the
// new timestamp; the Ikoria release notes: "Removing a keyword counter
// doesn't change the timestamp of any remaining counters." When the
// last one goes the entry goes with it.
//
// Keyword kinds only — a +1/+1 counter has no layer-6 meaning, and a
// stamp for it would be snapshot noise.
func (c *Card) stampKeywordCounter(kind string, delta, count int, now func() int64) {
	if !IsKeywordCounter(kind) {
		return
	}
	switch {
	case count <= 0:
		delete(c.CounterStampedAt, kind)
		if len(c.CounterStampedAt) == 0 {
			c.CounterStampedAt = nil
		}
	case delta > 0:
		if c.CounterStampedAt == nil {
			c.CounterStampedAt = make(map[string]int64, 1)
		}
		c.CounterStampedAt[kind] = now()
	}
}

// keywordCounterTimestamp is the CR 613.7c timestamp the layer pass
// orders one keyword counter kind at. A kind with no stamp — a restore
// point written before ADR 0101, or a fixture that writes Counters
// directly — is ordered at the permanent's own timestamp, the earliest
// moment the counter could have been put on it (ADR 0101 Decision 7,
// owner decision 4).
func (c *Card) keywordCounterTimestamp(kind string) int64 {
	if ts, ok := c.CounterStampedAt[kind]; ok {
		return ts
	}
	return c.layerTimestamp()
}

// keywordCounterEffect is one keyword counter kind on one permanent, as
// a CR 613.1f layer-6 continuous effect. Every counter of the kind
// shares the one CR 613.7c timestamp, so one effect carries them all.
type keywordCounterEffect struct {
	target  uuid.UUID
	keyword string
	// count is how many instances of the keyword the counters give:
	// keywordCounterInstances.
	count     int
	timestamp int64
}

func (e keywordCounterEffect) Layer() (Layer, SubLayer) { return Layer6Ability, 0 }
func (e keywordCounterEffect) Timestamp() int64         { return e.timestamp }
func (e keywordCounterEffect) AppliesTo(target *Card, _ *Game) bool {
	return target != nil && target.InstanceID == e.target
}
func (e keywordCounterEffect) Apply(c *Characteristic, _ *Card, _ *Game) {
	// Most of CR 122.1b's keywords are redundant when repeated (the
	// Ikoria release notes), and AppendKeywordAbility dedupes them: two
	// flying counters, or a flying counter on a creature that prints
	// flying, are one flying. Exalted is not: "A creature with multiple
	// exalted counters will have that many instances of exalted"
	// (Emissary of Soulfire ruling, 2024-06-07), so a cumulative kind
	// appends one instance per counter (keywordCounterInstances).
	for i := 0; i < max(e.count, 1); i++ {
		c.Abilities = AppendKeywordAbility(c.Abilities, e.keyword)
	}
}
func (e keywordCounterEffect) RemovesAbilities() bool      { return false }
func (e keywordCounterEffect) ContinuesAfterRemoval() bool { return false }

// keywordCounterEffectsLocked is the layer pass's source list for
// keyword counters: one effect per keyword kind on each battlefield
// permanent. Sorted by kind within a permanent, so two kinds stamped at
// the same instant fall in a stable order.
//
// Caller must hold g.mu in write mode (the layer pass does).
func (g *Game) keywordCounterEffectsLocked() []ContinuousEffect {
	if g.Battlefield == nil {
		return nil
	}
	var out []ContinuousEffect
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if len(c.Counters) == 0 {
			continue
		}
		for _, kind := range sortedCounterKinds(c.Counters) {
			if !IsKeywordCounter(kind) || c.Counters[kind] <= 0 {
				continue
			}
			out = append(out, keywordCounterEffect{
				target:    c.InstanceID,
				keyword:   kind,
				count:     keywordCounterInstances(kind, c.Counters[kind]),
				timestamp: c.keywordCounterTimestamp(kind),
			})
		}
	}
	return out
}

// keywordCounterTokens is the off-battlefield half of CR 122.1b: the
// keywords a card's own keyword counters give it, in sorted order, for
// a card no layer pass runs over. It is read by forEachAbilityToken
// (HasKeyword, protection) and by the view's off-battlefield ability
// list, so the badge and the rule cannot disagree.
func keywordCounterTokens(c *Card) []string {
	if c == nil || len(c.Counters) == 0 {
		return nil
	}
	var out []string
	for _, kind := range sortedCounterKinds(c.Counters) {
		if !IsKeywordCounter(kind) || c.Counters[kind] <= 0 {
			continue
		}
		for i := 0; i < keywordCounterInstances(kind, c.Counters[kind]); i++ {
			out = append(out, kind)
		}
	}
	return out
}

// keywordCounterInstances is how many instances of its keyword `n`
// counters of `kind` give: one per counter for a cumulative keyword
// (exalted, by the Emissary of Soulfire ruling of 2024-06-07), and one
// however many there are for every other kind (Decision 1). Both the
// layer pass and the off-battlefield walk read it, so the badge and the
// rule agree.
func keywordCounterInstances(kind string, n int) int {
	if n <= 0 {
		return 0
	}
	if KeywordIsCumulative(kind) {
		return n
	}
	return 1
}

// KeywordCounterTokens is keywordCounterTokens for the view package.
func KeywordCounterTokens(c Card) []string {
	return keywordCounterTokens(&c)
}

// copyStringInt64Map deep-copies a keyword-counter stamp map for an
// undo clone and the snapshot projections. nil in, nil out.
func copyStringInt64Map(in map[string]int64) map[string]int64 {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]int64, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
