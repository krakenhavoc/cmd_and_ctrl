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
// HasKeyword falls back to CatalogPrintedKeywords — the catalog-side
// hook populated by wire.go. The layer engine only maintains
// Effective() for battlefield cards, so flash (the only keyword that
// matters off the battlefield today) needs this fallback to gate a
// cast of an Ambush Viper from the caster's hand.
//
// Added in S18 sub-PR 2.

// HasKeyword reports whether the card has the named keyword. kw
// must be a canonical lowercase token (see AGENTS.md §7 "Adding a
// combat-keyword card" for the table): "flying", "reach",
// "first strike", "double strike", "deathtouch", "lifelink",
// "trample", "vigilance", "menace", "defender", "haste", "flash".
//
// On-battlefield: reads c.Effective().Abilities, so keywords granted
// by static abilities (Lord of Atlantis's islandwalk on other
// Merfolk) are included alongside the card's own printed keywords.
// Off-battlefield: falls back to the card's own `Keywords` (token
// templates) and then CatalogPrintedKeywords(c.OracleID), which
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
	for _, a := range CatalogPrintedKeywords(c.OracleID) {
		if a == kw {
			return true
		}
	}
	return false
}

// HasSummoningSickness reports whether the card is currently unable
// to attack or activate tap-cost abilities due to CR 302.1
// summoning sickness — entered the battlefield this turn AND does
// not have haste.
//
// Haste is a read-time bypass, NOT a clear-on-ETB: a creature that
// gains haste mid-turn (e.g. via Anger in graveyard, or a Concerted
// Effort-style grant) should be able to attack immediately that
// turn without waiting for next untap. Conversely, a creature that
// loses haste mid-turn (rare but possible via type-change effects)
// correctly becomes sick until next untap. Keeping
// `SummonedThisTurn` a pure "when did this creature enter" flag
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
// nil arguments return false (defensive — no card can block a
// missing attacker, and a nil blocker can't block).
func CanBlock(attacker, blocker *Card) bool {
	if attacker == nil || blocker == nil {
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
