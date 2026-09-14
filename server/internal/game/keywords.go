package game

// keywords.go is the single read surface for combat-keyword data.
// Every combat consumer in the engine (DeclareAttacker gate, block
// validation, damage substep collection, flash cast gate) goes
// through the helpers here instead of inline
// `for _, a := range card.Effective().Abilities` loops. Centralising
// the reader means one canonical tokeniser, one normalisation, and
// one place to fix a typo.
//
// Keywords are bare strings in Characteristic.Abilities. The S16
// layer engine already exposes them (Lord of Atlantis appends
// "islandwalk" in its Layer 6 Apply); the S18 `Spec.PrintedKeywords`
// catalog slot feeds the same slice via a synthesized layer-6
// StaticAbility at catalog load time, so a card's own printed
// keywords AND grants from other cards both land in
// Effective().Abilities on the battlefield.
//
// Off the battlefield (cards in hand, library, graveyard, exile),
// HasKeyword falls back to the card's own `Keywords` slice and then
// to CatalogPrintedKeywords — the catalog-side hook populated by
// wire.go. The layer engine only maintains Effective() for
// battlefield cards, so flash (the only keyword that matters off
// the battlefield today) needs this fallback to gate a cast of an
// Ambush Viper from the caster's hand — or, since #319 / #320, of
// any of the ~640 cards that print flash and have no catalog entry,
// from hand or from the command zone.
//
// Added in S18 sub-PR 2.

import (
	"strings"

	"github.com/google/uuid"
)

// canonicalKeywords is the set of keyword abilities this engine
// actually enforces — the S18 table, one entry per consumer in the
// combat / cast paths above, plus the S23 targeting pair (hexproof,
// shroud), whose consumer is the targeting choke point in
// targets.go, plus S25's indestructible, whose consumer is the
// destruction path (indestructible.go). The tokens are the canonical
// lowercase wire form; the client's KEYWORD_ICONS map is keyed by
// exactly these strings.
//
// The set is deliberately CLOSED. Scryfall publishes a `keywords`
// array carrying every mechanic printed on a card — "Prepared",
// "Waterbend", "Cycling", "Ward" — and the deck importer filters
// against this table before stamping Card.Keywords. Letting the
// rest through would put strings the engine can't act on into
// Characteristic.Abilities, where they would render as badges the
// player can't rely on and read as promises the rules layer never
// makes. A keyword joins this table in the same change that
// teaches the engine to honour it.
var canonicalKeywords = map[string]bool{
	"flying":        true,
	"reach":         true,
	"first strike":  true,
	"double strike": true,
	"deathtouch":    true,
	"lifelink":      true,
	"trample":       true,
	"vigilance":     true,
	"menace":        true,
	"defender":      true,
	"haste":         true,
	"flash":         true,
	"hexproof":      true,
	"shroud":        true,
	// S25 (#77): indestructible's consumer is the destruction path,
	// not combat or targeting — see indestructible.go for the rule
	// and for the list of things it deliberately does not stop.
	"indestructible": true,
	// changeling (CR 702.73) joins the table in S26, in the same
	// change that teaches Card.HasSubtype to honour it — which is
	// the rule the table's closedness encodes. Its consumer is not
	// in this file: it is HasAllCreatureTypes in creature_types.go,
	// and through it every "of the chosen type" / "shares a creature
	// type" read in the engine.
	KeywordChangeling: true,
}

// KeywordChangeling is the canonical token for changeling (CR
// 702.73a — "this card is every creature type"). Named because it is
// read from three packages and a typo in any of them would silently
// turn a changeling back into a Shapeshifter.
//
// It is the one keyword in the table whose meaning is not a combat or
// timing rule but a CHARACTERISTIC: it answers a question about the
// card's types, in every zone, and so it is consulted by HasSubtype
// rather than by the combat engine.
const KeywordChangeling = "changeling"

// CanBeTargetedBy reports whether `caster` may choose this card as
// the target of a spell or ability they control, under the CR 702
// protection-style keywords the engine honours:
//
//   - shroud (CR 702.18a) — can't be the target of spells or
//     abilities AT ALL, its own controller's included.
//   - hexproof (CR 702.11b) — can't be the target of spells or
//     abilities your OPPONENTS control.
//
// Two keywords that belong to the same family are deliberately
// ABSENT, and each is absent for a structural reason rather than
// for lack of time (see docs/decisions/0038-protection-style-keywords.md):
//
//   - ward — CR 702.21a makes ward a TRIGGERED ability, not a
//     targeting restriction. A warded permanent is a perfectly
//     legal target; the trigger then counters the spell unless its
//     controller pays. Enforcing it here would be the wrong rule in
//     the wrong place — it would refuse the target outright instead
//     of offering the payment.
//   - protection — CR 702.16e tests the quality against the SOURCE
//     of the spell or ability. This choke point receives only the
//     controller's ID and never the source object, so the test it
//     needs cannot be expressed here at all.
//
// Both are also PARAMETERISED keywords ("ward {2}", "protection
// from red") and Characteristic.Abilities is a []string of bare
// tokens, so there is nowhere to put the cost or the quality.
//
// Only the battlefield is checked. Both keywords are abilities of a
// permanent (CR 113.6 — a continuous effect from a static ability
// applies only while its source is on the battlefield, and the
// printed abilities name a permanent), so a creature card sitting in
// a graveyard that happens to print hexproof is a legal target for
// Regrowth. Without this zone guard the off-battlefield HasKeyword
// fallback — which reads the card's own printed Keywords — would
// wrongly protect it there.
//
// Players are not covered: hexproof on a PLAYER (Leyline of
// Sanctity) has no home yet, because Player carries no keyword
// slice. TargetPlayer refs pass this gate by not reaching it.
//
// nil card returns true: a caller that has lost the card has an
// existence problem, not a targeting one, and the zone walk that
// wraps this reports that separately.
func CanBeTargetedBy(c *Card, zone ZoneKind, caster uuid.UUID) bool {
	if c == nil || zone != ZoneBattlefield {
		return true
	}
	if HasKeyword(c, "shroud") {
		return false
	}
	if HasKeyword(c, "hexproof") && c.Controller != caster {
		return false
	}
	return true
}

// CanonicalKeyword normalises one printed keyword string to the
// engine's wire token, reporting whether the engine knows it.
// Scryfall capitalises its keyword arrays as the card prints them
// ("Flying", "First strike", "Double strike"), so normalisation is
// a lowercase plus a whitespace trim; anything outside
// canonicalKeywords returns ("", false).
//
// Exported for the deck importer, which is the only caller —
// keeping the table in this package means the reader
// (HasKeyword) and the writer (deck.printedKeywords) can never
// drift on spelling. Added with the #317 / #319 / #320 fix.
func CanonicalKeyword(s string) (string, bool) {
	kw := strings.ToLower(strings.TrimSpace(s))
	if !canonicalKeywords[kw] {
		return "", false
	}
	return kw, true
}

// HasKeyword reports whether the card has the named keyword. kw
// must be a canonical lowercase token (see AGENTS.md §7 "Adding a
// combat-keyword card" for the table): "flying", "reach",
// "first strike", "double strike", "deathtouch", "lifelink",
// "trample", "vigilance", "menace", "defender", "haste", "flash",
// "hexproof", "shroud", "indestructible".
//
// On-battlefield: reads c.Effective().Abilities, so keywords granted
// by static abilities (Lord of Atlantis's islandwalk on other
// Merfolk) are included alongside the card's own printed keywords.
// Off-battlefield: falls back to the card's own `Keywords` (every
// deck-imported card since #317 / #319 / #320; token templates
// before that) and then CatalogPrintedKeywords(c.OracleID), which
// returns the `Spec.PrintedKeywords` slot. Granted keywords
// don't apply off the battlefield (CR 113.6 — continuous effects
// from static abilities only apply while the source permanent is on
// the battlefield), so the fallback is correct.
//
// nil card returns false (makes caller code that handles missing
// lookups terser).
func HasKeyword(c *Card, kw string) bool {
	if c == nil || kw == "" {
		return false
	}
	// Battlefield path: the layer engine populates `effective` on
	// entry. The `EnteredBattlefieldAt > 0` guard is a defensive
	// check for cards that have been on the battlefield in this
	// game's lifetime but briefly have a stale/nil cache; in
	// practice the recompute pass ensures `effective` is non-nil
	// on battlefield cards before any consumer reads it.
	if c.effective != nil {
		for _, a := range c.effective.Abilities {
			if a == kw {
				return true
			}
		}
		return false
	}
	// Off-battlefield path: the card's own printed keywords first
	// (S21 sub-PR 1 — token templates carry them on the Card, since
	// a token has no oracle ID for the catalog to key on), then the
	// catalog. Lookup-miss (non-catalog card) returns nil → no
	// keywords → false.
	for _, a := range c.Keywords {
		if a == kw {
			return true
		}
	}
	if CatalogPrintedKeywords == nil || c.OracleID == "" {
		return false
	}
	for _, a := range CatalogPrintedKeywords(CatalogKey(*c)) {
		if a == kw {
			return true
		}
	}
	return false
}

// HasSummoningSickness reports whether the card is currently unable
// to attack or to activate an ability with {T} or {Q} in its cost,
// per CR 302.6 — it is a CREATURE, it entered the battlefield this
// turn, and it does not have haste.
//
// The creature test is part of the rule, not a caller's garnish.
// CR 302.6 is written about creatures and nothing else: a Treasure
// token, a Sol Ring and a Fabled Passage are not creatures, and
// their controller may tap them the turn they arrive. #530: this
// helper used to answer on SummonedThisTurn alone, which made every
// mana rock, fetchland and Treasure played this turn read as
// unusable. The engine's write paths each carried their own
// `card.IsCreature() &&` prefix and so stayed correct, but the wire
// (protocol/view.go) called the helper bare and shipped
// `summoning_sick: true` for every permanent that entered this
// turn, so the client greyed abilities the server would have
// allowed — the two player reports #365 and #368. Those `IsCreature`
// prefixes are now redundant rather than load-bearing; they are
// left in place as documentation at the point of use.
//
// Creature-hood is read at call time, which is what CR 302.6 asks
// for: the clock runs on continuous CONTROL since the turn began,
// not on when the permanent became a creature. So a Vehicle crewed
// this turn but on the battlefield since last turn is not sick and
// can attack, while one that landed this turn is sick however early
// it was crewed. Same for a manland animated the turn it entered.
// Callers reading types after a state change must have called
// RecomputeLayersIfStaleLocked, exactly as for CurrentPower.
//
// Haste is a read-time bypass, NOT a clear-on-ETB: a creature that
// gains haste mid-turn (e.g. via Anger in graveyard, or a Concerted
// Effort-style grant) should be able to attack immediately that
// turn without waiting for next untap. Conversely, a creature that
// loses haste mid-turn (rare but possible via type-change effects)
// correctly becomes sick until next untap. Keeping
// `SummonedThisTurn` a pure "when did this permanent enter" flag
// decouples the two independent state changes.
//
// nil card returns false.
func HasSummoningSickness(c *Card) bool {
	if c == nil {
		return false
	}
	if !c.SummonedThisTurn {
		return false
	}
	// CR 302.6 gates creatures. Everything else is free the turn it
	// arrives.
	if !c.IsCreature() {
		return false
	}
	return !HasKeyword(c, "haste")
}

// CanBlock reports whether blocker can legally block attacker given
// evasion keywords (flying, menace, fear, shadow, etc.) AND
// defensive keywords (reach, reach-adjacent). Does NOT check tap
// state, declare-blockers-step gating, or "defender can't attack"
// — those are caller-side checks specific to the declaration path.
// A creature with defender CAN block; defender only restricts
// attacking.
//
// S18 rules enforced:
//   - Flying attackers can only be blocked by flying or reach
//     blockers (CR 702.9b).
//   - Menace is not checked here — it's a block-count rule, see
//     BlockerCountValid. This function runs per-pair.
//
// S24 adds the restriction vocabulary (restrictions.go), and this is
// the ONE place both halves of it are read — "~ can't block" on the
// blocker (Pacifism, Carrion Feeder) and "~ can't be blocked" on the
// attacker (Whispersilk Cloak, Rogue's Passage). Putting both here
// rather than splitting the blocker's half into BlockerEligible is
// deliberate: DeclareBlocker calls only this function, while the
// legal-move enumerator and the #328 auto-pass signal call
// BlockerEligible AND this function. One predicate all three reach is
// the only arrangement in which the engine cannot refuse a block the
// enumerator offered.
//
// nil arguments return false (defensive — no card can block a
// missing attacker, and a nil blocker can't block).
func CanBlock(attacker, blocker *Card) bool {
	if attacker == nil || blocker == nil {
		return false
	}
	// CR 509.1b restrictions. Read before the evasion keywords
	// because they are absolute: no defensive keyword answers
	// "can't be blocked" the way reach answers flying.
	if Restricted(attacker, CantBeBlocked) {
		return false
	}
	if Restricted(blocker, CantBlock) {
		return false
	}
	// Flying: blocker must have flying or reach.
	if HasKeyword(attacker, "flying") {
		if !HasKeyword(blocker, "flying") && !HasKeyword(blocker, "reach") {
			return false
		}
	}
	return true
}

// BlockerCountValid reports whether the given blocker set is
// legal against the attacker under block-count keywords (menace).
// Called at the declare-blockers step close-out, per CR 702.110.
//
// Menace requires ≥2 blockers: a single blocker against a menace
// attacker is illegal and the single block is reverted (attacker
// becomes unblocked). An empty blockers list is always legal —
// "no blockers" is always a valid outcome.
//
// nil attacker returns true (no keyword to enforce). Nil entries
// in the blockers slice are counted as blockers (the caller has
// validated the list; defensive filtering would swallow real
// state bugs).
func BlockerCountValid(attacker *Card, blockers []*Card) bool {
	if attacker == nil {
		return true
	}
	if len(blockers) == 0 {
		return true
	}
	if HasKeyword(attacker, "menace") && len(blockers) < 2 {
		return false
	}
	return true
}
