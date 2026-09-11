package game

// characteristic.go is the S16 layer-system foundation: a mutable
// snapshot of a card's effective characteristics post-layer
// resolution. The wire ships effective values; the engine's rules
// logic that needs printed values (mana cost, owner) reads off the
// Card directly.
//
// Characteristic is INTENTIONALLY a flat struct of value types — the
// layer engine recomputes from scratch on every relevant event
// (XMage / Forge style) so cheap copying matters more than sharing.
// One Characteristic per battlefield card per recompute cycle.
//
// Sub-PR 1 ships the type + the printed-characteristic accessor;
// sub-PR 3 wires the layer engine to populate it. Until then
// Card.Effective() returns the printed characteristic verbatim — the
// wire shape is unchanged.

// Characteristic is the post-layer-resolution view of a single card's
// gameplay-relevant fields. Mirrors the subset of Card fields that
// continuous effects can change (CR 613 layers 1-7); does NOT carry
// fields that are immutable per-instance (Owner, InstanceID,
// ScryfallID) or that belong to other engine subsystems (Counters
// is on Card; Tapped is on Card).
//
// Power / Toughness / Types / Subtypes / Supertypes / Colors /
// Abilities / Name are the only fields that get layered. New
// layer-touching fields land here as the catalog grows. Loyalty is
// NOT here — it lives in Card.Counters["loyalty"] as a counter, not
// a layered characteristic (CR 121 / 613 distinction).
type Characteristic struct {
	Power      int
	Toughness  int
	Types      []string
	Subtypes   []string
	Supertypes []string
	Colors     []string
	Abilities  []string
	Name       string
}

// printedCharacteristic builds a Characteristic from the card's
// immutable printed fields. Called by the layer engine as the
// starting point for every recompute pass — printed values are the
// "layer 0" baseline that layers 1..7 mutate in order.
//
// Type-line parsing is intentionally simple: split on the em-dash
// separator ("Legendary Creature — Human Wizard"), take the left
// half as supertypes + types, the right half as subtypes. Supertypes
// are recognised by name from the small fixed list (legendary,
// basic, snow, world, ongoing). Everything else on the left is a
// type. Empty type lines (placeholder demo cards) yield a
// Characteristic with empty type slices — safe for downstream
// layers to iterate.
//
// Colors are NOT derived here — there's no per-card color field on
// the Card struct yet (S15 added ManaCost only). When S16's catalog
// adds color-changing effects, this will need a parsed-from-cost
// initial color set.
func (c Card) printedCharacteristic() Characteristic {
	supertypes, types, subtypes := ParseTypeLine(c.TypeLine)
	// Printed keywords come from two places. The catalog's
	// Spec.PrintedKeywords slot (S18 sub-PR 2) is the older one;
	// including it here means off-battlefield
	// CardView.Abilities surfaces the keyword on hand cards — the
	// client's cast-timing gate needs flash to grey-enable Ambush
	// Viper at instant speed. The on-battlefield synth adds these
	// via a Layer 6 StaticAbility with a dedupe, so double-counting
	// is impossible.
	//
	// Card.Keywords is the printed-data road: the deck importer
	// stamps Scryfall's `keywords` array onto every imported card
	// (#317 / #319 / #320), and token templates declare theirs
	// inline because a token has no oracle ID for the catalog hook
	// to key on (S21 sub-PR 1). The catalog remains a fallback and
	// an override for cards that never go through deck import —
	// fixtures, tokens, and any spec that deliberately states a
	// keyword Scryfall doesn't.
	//
	// The two sources overlap for every catalog card that is also
	// imported from a decklist, so the merge dedupes: a doubled
	// "flash" is harmless to HasKeyword but renders as two badges
	// on the client's keyword row.
	var abilities []string
	if CatalogPrintedKeywords != nil && c.OracleID != "" {
		if kws := CatalogPrintedKeywords(c.OracleID); len(kws) > 0 {
			abilities = append(abilities, kws...)
		}
	}
	for _, kw := range c.Keywords {
		if !containsKeyword(abilities, kw) {
			abilities = append(abilities, kw)
		}
	}
	return Characteristic{
		Power:      c.Power,
		Toughness:  c.Toughness,
		Types:      types,
		Subtypes:   subtypes,
		Supertypes: supertypes,
		Colors:     printedColorsFromCost(c.ManaCost),
		Name:       c.Name,
		Abilities:  abilities,
	}
}

// containsKeyword reports whether xs already holds kw. Used by the
// printed-characteristic merge; the battlefield's Layer 6 keyword
// synth keeps its own copy over in cards/effects, where the name
// would otherwise collide with that package's test helpers.
func containsKeyword(xs []string, kw string) bool {
	for _, x := range xs {
		if x == kw {
			return true
		}
	}
	return false
}

// printedColorsFromCost extracts the unique WUBRG letters from a
// printed mana-cost string. Drives the printed Colors slice on
// Characteristic, which feeds commander-identity computation
// (S16 sub-PR 5) and any future Layer 5 color-changing effect.
//
// Pure cost-based; doesn't read color-indicator stamps or rules-
// text color words (Bant Charm-style cards). The CR 105.2c "color
// indicator" + CR 903.4 "rules-text color" cases are sandbox
// simplifications — the layer engine's Effective().Colors stays
// the canonical surface so a future card that mutates color works
// through the same path.
func printedColorsFromCost(cost string) []string {
	if cost == "" {
		return nil
	}
	seen := map[byte]bool{}
	out := []string{}
	for i := 0; i < len(cost); i++ {
		b := cost[i]
		if b >= 'a' && b <= 'z' {
			b -= 'a' - 'A'
		}
		if b == 'W' || b == 'U' || b == 'B' || b == 'R' || b == 'G' {
			if !seen[b] {
				seen[b] = true
				out = append(out, string(b))
			}
		}
	}
	return out
}

// Effective returns the card's post-layer-resolution characteristic.
// Sub-PR 3: reads `Card.effective` when populated by the layer
// engine; falls back to printedCharacteristic when nil (card has
// either never been on the battlefield since the most recent
// recompute or is currently in another zone).
//
// Returns by value so callers can't mutate the cache. The layer
// engine writes through a different path (recompute owns the
// pointer it allocates per cycle and replaces atomically).
func (c Card) Effective() Characteristic {
	if c.effective != nil {
		return *c.effective
	}
	return c.printedCharacteristic()
}

// ParseTypeLine splits a Scryfall-style type line into supertypes,
// types, and subtypes. Examples:
//
//	"Legendary Creature — Human Wizard"
//	  → supertypes: ["Legendary"], types: ["Creature"],
//	    subtypes: ["Human", "Wizard"]
//	"Basic Land — Forest"
//	  → supertypes: ["Basic"], types: ["Land"], subtypes: ["Forest"]
//	"Sorcery"
//	  → supertypes: nil, types: ["Sorcery"], subtypes: nil
//	""
//	  → all nil
//
// The em-dash "—" (U+2014) is the canonical Scryfall separator;
// hyphen-minus "-" is also accepted as a fallback for cards whose
// importer round-tripped through ASCII. Whitespace tokenisation
// handles double-spaces around the dash.
//
// Exported (S16 sub-PR 4) so the protocol layer can do the printed-
// vs-effective comparison in effectiveTypeLine without re-implementing
// the parser.
func ParseTypeLine(typeLine string) (supertypes, types, subtypes []string) {
	if typeLine == "" {
		return nil, nil, nil
	}
	left, right := typeLine, ""
	for i := 0; i < len(typeLine); i++ {
		// Look for the em-dash (3-byte UTF-8 sequence E2 80 94) or
		// a plain hyphen surrounded by spaces.
		if i+2 < len(typeLine) && typeLine[i] == 0xE2 && typeLine[i+1] == 0x80 && typeLine[i+2] == 0x94 {
			left = typeLine[:i]
			right = typeLine[i+3:]
			break
		}
	}
	leftTokens := splitWords(left)
	for _, tok := range leftTokens {
		if isSupertype(tok) {
			supertypes = append(supertypes, tok)
		} else {
			types = append(types, tok)
		}
	}
	subtypes = splitWords(right)
	return supertypes, types, subtypes
}

// splitWords tokenises on runs of whitespace. Empty input → nil.
func splitWords(s string) []string {
	var out []string
	start := -1
	for i := 0; i < len(s); i++ {
		if s[i] == ' ' || s[i] == '\t' || s[i] == '\n' {
			if start >= 0 {
				out = append(out, s[start:i])
				start = -1
			}
		} else if start < 0 {
			start = i
		}
	}
	if start >= 0 {
		out = append(out, s[start:])
	}
	return out
}

// isSupertype reports whether `s` names a CR 205.4 supertype.
// Case-sensitive: Scryfall's type lines normalise capitalisation.
func isSupertype(s string) bool {
	switch s {
	case "Legendary", "Basic", "Snow", "World", "Ongoing", "Token", "Tribal":
		return true
	}
	return false
}
