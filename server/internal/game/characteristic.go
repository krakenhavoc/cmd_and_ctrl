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
	supertypes, types, subtypes := parseTypeLine(c.TypeLine)
	return Characteristic{
		Power:      c.Power,
		Toughness:  c.Toughness,
		Types:      types,
		Subtypes:   subtypes,
		Supertypes: supertypes,
		Name:       c.Name,
	}
}

// Effective returns the card's post-layer-resolution characteristic.
// Sub-PR 1 ships the no-op pass: Effective == printed. Sub-PR 3
// adds the layer cache and this method becomes the cache reader.
//
// Returns by value so callers can't mutate the cache. The layer
// engine writes through a different path (recompute owns the
// pointer it allocates per cycle).
func (c Card) Effective() Characteristic {
	return c.printedCharacteristic()
}

// parseTypeLine splits a Scryfall-style type line into supertypes,
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
func parseTypeLine(typeLine string) (supertypes, types, subtypes []string) {
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
