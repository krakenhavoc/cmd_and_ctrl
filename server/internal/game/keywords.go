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

import "strings"

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
	// Landwalk (CR 702.14) joins with #705, in the same change that
	// teaches BlockPairRefusalLocked to read the defending player's
	// lands (landwalk.go). One token per kind the engine enforces;
	// the rarer variants (snow swampwalk, legendary landwalk,
	// desertwalk) join with their first card. landwalkTokens in
	// landwalk.go is the other half of this list, and
	// TestLandwalkTokensAreCanonical keeps the two in step.
	"plainswalk":        true,
	"islandwalk":        true,
	"swampwalk":         true,
	"mountainwalk":      true,
	"forestwalk":        true,
	"nonbasic landwalk": true,
	"fear":              true,
	"intimidate":        true,
	"shadow":            true,
	"horsemanship":      true,
	"skulk":             true,
	// cycling (CR 702.29) joins with #660, in the same change that
	// teaches the engine to honour it: a cycling ability is a real
	// CR 602 activation from hand now (effects.Cycling,
	// ActivatedAbilityShape.Zones), and cycling a card emits
	// EventCycle. Unlike the combat keywords above it has no
	// consumer IN the engine's rules paths — the ability is on the
	// card, not in a table — so the token's job is the badge and
	// the ADR 0037 coverage signal, which is exactly what it was
	// being withheld for.
	//
	// Typecycling's printed keywords ("basic landcycling",
	// "plainscycling", "slivercycling") are separate Scryfall tokens
	// and stay OUT. They are one per creature type and one per land
	// type — dozens of strings for a badge that would say less than
	// the cycling row the hand card now shows — and the one card
	// that prints one today (Sylvan Reclamation) is an instant,
	// which has no badge row at all.
	"cycling": true,
	// protection (CR 702.16) joins with #662, in the same change that
	// teaches all four of its DEBT checks to read it — targeting
	// (CanBeTargetedBy, below), attachment (attach.go), damage
	// (builtin_replacements.go) and blocking (block_legality.go).
	//
	// It is the table's one PARAMETERISED entry, and so the one entry
	// that is not itself a wire token. The tokens are "protection
	// from <quality>" and they are minted by ProtectionTokens
	// (protection.go), which is also the only thing that parses them;
	// CanonicalKeywords refuses a bare "protection", because a badge
	// that does not say what it protects from is a promise the rules
	// layer cannot keep. See docs/decisions/0072-protection.md §1.
	KeywordProtection: true,
	// foretell (CR 702.143) joins with #658, in the same change that
	// teaches the engine to honour it: foretelling is a real CR 116.2
	// special action now (game.PerformSpecialAction), the exiled card
	// is a CR 702.143b face-down object its owner may look at, and
	// the cast out of exile rides a per-instance CastPermission
	// priced at the foretell cost.
	//
	// Like cycling, its consumer is not a table in this file — the
	// keyword is on the card, not in the combat or targeting paths —
	// so the token's job is the badge and the ADR 0037 coverage
	// signal. What it must NOT do is join before the mechanic works:
	// a badge promising foretell on a card the engine can only
	// hard-cast is the half-a-card failure ADR 0037 §5 forbids.
	"foretell": true,
	// suspend (CR 702.62) joins with #659, in the same change that
	// teaches the engine to honour it: the special action, the time
	// counters, the exile-zone upkeep countdown (#925's
	// TriggeredAbility.Zones), the free cast when the last one comes
	// off and CR 702.62e's haste. Same reasoning as foretell above —
	// the consumer is the card, not a table in this file, so the
	// token's job is the badge and the ADR 0037 coverage signal, and
	// it must not arrive before the mechanic does.
	"suspend": true,
	// madness (CR 702.35) joins with #657, in the same change that
	// teaches the engine to honour it: the discard replacement over
	// #650's one discard event, the exile-zone trigger that offers
	// the cast (#925's Zones dimension), and the per-instance
	// CastPermission keyed "madness" with TimingFlash that carries
	// CR 608.2g. Same reasoning as foretell and suspend above — the
	// consumer is the card, not a table in this file, so the token's
	// job is the badge and the ADR 0037 coverage signal, and it must
	// not arrive before the mechanic does.
	"madness": true,
	// phasing (CR 702.26) joins with #1199, in the same change that
	// teaches the engine to honour it — which is the closedness rule
	// this table states. Unlike foretell, suspend and madness above,
	// its consumer IS in the engine and is one function:
	// performPhasingLocked (phasing.go) reads it off the effective
	// characteristics to decide which of the active player's
	// permanents phase out during CR 502.1's turn-based action.
	//
	// Reading it off Effective().Abilities rather than off Keywords is
	// what makes Shimmer's "each land of the chosen type has phasing"
	// and Vanishing's aura work like a printed one, and what makes a
	// permanent that has lost all abilities stop phasing.
	//
	// Stamped by the deck importer like every other canonical token,
	// so the ~70 printed-phasing permanents of the Mirage and Visions
	// cycle are correct with no catalog entry at all. See
	// KeywordPhasing in phasing.go and ADR 0084.
	KeywordPhasing: true,
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

// CanBeTargetedBy reports whether the spell or ability `src`
// describes may choose this card as a target, under the CR 702
// protection-style keywords the engine honours:
//
//   - shroud (CR 702.18a) — can't be the target of spells or
//     abilities AT ALL, its own controller's included.
//   - hexproof (CR 702.11b) — can't be the target of spells or
//     abilities your OPPONENTS control.
//   - protection (CR 702.16b) — can't be the target of spells with
//     the stated quality, or of abilities from a SOURCE with it.
//
// The three are not the same shape, and that is the whole reason
// this function takes a TargetSource rather than a caster ID.
// Hexproof asks who is casting; protection asks what is casting. A
// white player's Lightning Bolt is red; a red player's Sol Ring
// ability is colourless. #662 threaded the source through every
// targeting call site so the second question could be asked here.
//
// Ward is still deliberately ABSENT, for the reason ADR 0038 §7
// gives and #95 confirmed by shipping it elsewhere: CR 702.21a makes
// ward a TRIGGERED ability, not a targeting restriction. A warded
// permanent is a perfectly legal target; the trigger then counters
// the spell unless its controller pays. Enforcing it here would
// refuse the target outright instead of offering the payment. It
// lives in cards/effects/ward.go.
//
// Only the battlefield is checked. All three keywords are abilities
// of a permanent (CR 113.6 — a continuous effect from a static
// ability applies only while its source is on the battlefield, and
// the printed abilities name a permanent), so a creature card
// sitting in a graveyard that happens to print hexproof is a legal
// target for Regrowth. Without this zone guard the off-battlefield
// ability fallback — which reads the card's own printed Keywords —
// would wrongly protect it there.
//
// Players are not covered HERE, and since #1197 that is a split
// rather than a gap: hexproof and protection on a PLAYER (Leyline of
// Sanctity, Teferi's Protection) are answered by
// canPlayerBeTargetedByLocked in player_statics.go, at the same two
// targeting call sites and on the same `targeting` switch. A
// TargetPlayer ref goes there; a TargetCard ref comes here. Two
// stores, because a player has no Characteristic and no layer — one
// rule, because both read the same protection grammar.
//
// nil card returns true: a caller that has lost the card has an
// existence problem, not a targeting one, and the zone walk that
// wraps this reports that separately.
func CanBeTargetedBy(c *Card, zone ZoneKind, src TargetSource) bool {
	if c == nil || zone != ZoneBattlefield {
		return true
	}
	if HasKeyword(c, "shroud") {
		return false
	}
	if HasKeyword(c, "hexproof") && c.Controller != src.Controller {
		return false
	}
	// CR 702.16b. The pre-test is cheap and the snapshot is not, so
	// a board with no protection on it pays one ability-list walk.
	if HasProtection(c) && ProtectedFrom(c, src.Characteristics()) {
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
//
// One printed clause can be more than one ability — "protection from
// Demons and from Dragons" is two (CR 702.16m) — so a caller
// scanning printed TEXT wants CanonicalKeywords below. This singular
// form answers false for such a clause rather than picking one of
// the two.
func CanonicalKeyword(s string) (string, bool) {
	kws, ok := CanonicalKeywords(s)
	if !ok || len(kws) != 1 {
		return "", false
	}
	return kws[0], true
}

// CanonicalKeywords is the plural form: the engine tokens one printed
// keyword-ability clause stands for, or ("", false) for a clause the
// engine does not enforce.
//
// Protection (CR 702.16) is the only keyword that can yield more than
// one token, and the only one whose token carries a parameter. The
// three readers that scan printed TEXT — deck.printedKeywords,
// deck.keywordLines and coverage.keywordOnlyLine — call this, so what
// the importer stamps and what the ADR 0037 coverage signal believes
// can never drift. A quality protection.go's closed grammar cannot
// parse yields NOTHING, which leaves the card flagged unimplemented
// and errs weaker.
//
// A bare "Protection" is refused for the same reason: Scryfall's
// keywords array carries it on every protection card and the quality
// lives only in the oracle text, so taking the array at its word
// would stamp a badge with nothing behind it (ADR 0038 §6 made the
// same call for "Hexproof from").
func CanonicalKeywords(s string) ([]string, bool) {
	kw := strings.ToLower(strings.TrimSpace(s))
	if kw == KeywordProtection {
		return nil, false
	}
	if canonicalKeywords[kw] {
		return []string{kw}, true
	}
	if toks, ok := ProtectionTokens(s); ok {
		return toks, true
	}
	return nil, false
}

// HasKeyword reports whether the card has the named keyword. kw
// must be a canonical lowercase token (see AGENTS.md §7 "Adding a
// combat-keyword card" for the table): "flying", "reach",
// "first strike", "double strike", "deathtouch", "lifelink",
// "trample", "vigilance", "menace", "defender", "haste", "flash",
// "hexproof", "shroud", "indestructible", "changeling", fear,
// intimidate, shadow, horsemanship, skulk, and the landwalk tokens
// ("islandwalk", "nonbasic landwalk", …).
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
	found := false
	forEachAbilityToken(c, func(a string) bool {
		if a == kw {
			found = true
			return false
		}
		return true
	})
	return found
}

// forEachAbilityToken walks the card's ability tokens in the zone
// rules HasKeyword documents, calling fn until it returns false.
// Allocation-free, because it is on the targeting and combat hot
// paths.
//
// Factored out of HasKeyword for ProtectionQualities (#662), which
// asks the same question of the same list and must not grow a second
// copy of the zone rules: a protection granted by an Equipment's
// layer-6 static and a protection the card prints have to answer
// alike, in every zone, or the four DEBT checks disagree with each
// other.
//
// nil card walks nothing.
func forEachAbilityToken(c *Card, fn func(token string) bool) {
	if c == nil {
		return
	}
	// Battlefield path: the layer engine populates `effective` on
	// entry. In practice the recompute pass ensures `effective` is
	// non-nil on battlefield cards before any consumer reads it.
	if c.effective != nil {
		for _, a := range c.effective.Abilities {
			if !fn(a) {
				return
			}
		}
		return
	}
	// CR 708.2a, ADR 0069: a face-down permanent has no text and so
	// no keywords. CatalogKey already answers "" for it, but
	// Card.Keywords is the deck importer's own road and bypasses the
	// catalog entirely — without this guard a face-down Ambush Viper
	// would still have flash.
	if c.FaceDownIsPermanent() {
		return
	}
	// Off-battlefield path: the card's own printed keywords first
	// (S21 sub-PR 1 — a token template's keywords are plain data on
	// the Card and always were), then the catalog. Lookup-miss
	// (non-catalog card) returns nil → no keywords.
	for _, a := range c.Keywords {
		if !fn(a) {
			return
		}
	}
	if CatalogPrintedKeywords == nil {
		return
	}
	// The empty KEY is the skip, not an empty oracle ID (ADR 0083
	// decision 3). A token reaches the catalog under its token key
	// like every other object; no token template declares
	// PrintedKeywords today — its keywords ride the Card above — so
	// this is the gate asking the right question rather than a
	// behaviour change.
	key := CatalogKey(*c)
	if key == "" {
		return
	}
	for _, a := range CatalogPrintedKeywords(key) {
		if !fn(a) {
			return
		}
	}
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
