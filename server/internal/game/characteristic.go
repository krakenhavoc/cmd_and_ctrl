package game

import "github.com/google/uuid"

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

	// AllCreatureTypes records that this object is every creature
	// type (CR 702.73a). It is a LAYER 4 TYPE FACT and it lives here,
	// next to Subtypes, rather than as the `changeling` keyword in
	// Abilities, which is where it used to live.
	//
	// The keyword was only ever storage — "every creature type" as
	// ~345 entries in Subtypes would make the wire type line
	// unreadable and every subtype loop quadratic — but storage in
	// the ability list leaked into the rules answer in both
	// directions (#670):
	//
	//   - a layer-6 ability removal emptied Abilities and with it
	//     deleted a layer-4 type fact, which is exactly what
	//     Maskwood Nexus' 2021-02-05 ruling says must NOT happen
	//     ("If an effect causes a creature with changeling to lose
	//     all abilities, it will remain all creature types … because
	//     changeling applies before the effect that removes it");
	//   - a later layer-4 subtype SET left the keyword behind, so a
	//     Kenrith's Transformation newer than the Nexus produced an
	//     Elk that was still every creature type.
	//
	// Both fall out for free now: the flag is cleared and re-set
	// inside layer 4 in timestamp order (SetSubtypes clears it, a
	// grant sets it), and layer 6 cannot see it.
	//
	// Printed changeling seeds it in printedCharacteristic — CR
	// 702.73a is a characteristic-defining ability, so CR 613.2
	// applies it before any other layer-4 effect, which is what
	// seeding the layer-0 baseline means in this engine. The keyword
	// stays in Abilities as the PRINTED source of the flag and as the
	// client's badge; nothing but the projection reads it.
	AllCreatureTypes bool

	// AbilitiesRemoved records that a CR 613.1f ability-removing
	// continuous effect ("loses all abilities", "is a colorless
	// Forest land") applied to this object in layer 6.
	//
	// It is NOT derivable from `Abilities` being empty. That slice
	// holds keywords, and a vanilla bear has none of those while
	// still having every activated, triggered, mana and static
	// ability its card prints. The engine reads printed abilities
	// out of the catalog at USE time through the Catalog* hooks, so
	// the layer engine needs somewhere to say "stop asking" — this
	// is that place, and game.CatalogAbilityKey is the accessor that
	// reads it.
	//
	// Timestamps (CR 613.6) do not need a second field. Printed
	// abilities are part of the object and are removed by any
	// removal effect that applies to it, whenever either arrived;
	// granted KEYWORDS ride in `Abilities` and are governed by the
	// layer-6 bucket's timestamp sort, which clears the slice in the
	// removal's slot and lets a later grant append after it.
	//
	// Since ADR 0093 it means "its OWN abilities are gone" — printed,
	// copy-granted (CR 707.9a) and token-template — and not "it has no
	// abilities": a layer-6 grant sorted after the removal survives in
	// GrantedAbilities below, and every reader keeps that half.
	AbilitiesRemoved bool

	// GrantedAbilities are the abilities OTHER effects gave this
	// object in layer 6 (CR 113.10, CR 613.1f) — Cryptolith Rite's
	// "{T}: Add one mana of any color" on every creature you control,
	// Chromatic Lantern's on every land. ADR 0093 Decision 1.
	//
	// Each entry names a catalog bundle by key, never a closure: a
	// grant is static catalog data (effects.AbilityGrant), registered
	// under GrantKey, and the reader that wants the abilities asks the
	// catalog for that key exactly as it asks for a card's own. So the
	// last-known-information copy of this struct (the dies harvest)
	// and a snapshot of it both carry nothing but strings and IDs.
	//
	// Written only by GrantAbility, from a layer-6 effect's Apply, in
	// that effect's timestamp slot — the order of this slice IS the
	// layer-6 order. A CR 613.1f removal empties it in its own slot,
	// which is what lets a grant with a LATER timestamp survive the
	// removal (CR 613.6), exactly as a granted keyword does in
	// Abilities above.
	//
	// NOT copiable (CR 707.2): a copy effect reads PrintedValues, never
	// the layered result, so nothing has to remember to leave these
	// out. The copiable grant is Card.GrantedAbilities, a different
	// field on a different struct (#665).
	GrantedAbilities []GrantedAbility

	// Controller is the post-layer-2 controller (CR 613.1b). It is
	// the one field here that is NOT a characteristic in the CR 109.3
	// sense — control is a property of the object, not of its
	// printed face — and it lives here anyway because layer 2 is a
	// layer, it has to sort by timestamp against every other layer,
	// and there was nowhere else for its output to land.
	//
	// Seeded from Card.baseController(), overwritten by any layer-2
	// effect that applies, and then MATERIALISED back onto
	// Card.Controller at the end of the recompute. That last step is
	// what makes this field worth having: ~385 sites in the engine
	// and the catalog read Card.Controller, and materialising means
	// all of them are right without being touched. See
	// recomputeLayersLocked.
	Controller uuid.UUID

	// ControlSource names the card whose layer-2 effect last wrote
	// Controller in this pass — the Act of Treason, the Mind Control,
	// the Switcheroo. Zero when no layer-2 effect applied, which is
	// the ordinary case AND the one that matters: a permanent going
	// home because the effect that took it ended has no source, and
	// the EventControlChanged materialiseControlLocked emits says so
	// by leaving Event.Source nil (#930).
	//
	// Written by the layer-2 Apply closures themselves (the engine's
	// controlStatic, the catalog's ControlAttachedBySource), so the
	// last write in the bucket's timestamp order is the effect that
	// won CR 613.7 — the same rule that decided Controller, read off
	// the same assignment rather than re-derived.
	ControlSource uuid.UUID

	// PTDefined records that a layer 7a or 7b effect SET this
	// object's power and toughness in this pass — a
	// characteristic-defining ability (CR 613.4a, Tarmogoyf,
	// Consuming Aberration) or an effect that sets P/T to specific
	// values (CR 613.4b, Humility, Hallowed Haunting over its own
	// Spirit tokens).
	//
	// It is engine bookkeeping rather than a characteristic, and it
	// exists for one reader: the toughness state-based action's
	// stand-in skip (Card.ToughnessIsKnown). The importer writes a 0
	// into Card.Toughness for a printed `*`, and the skip is what
	// stops CR 704.5f from eating every `*` creature whose
	// characteristic-defining ability this engine does not compute.
	// When the engine DOES compute it, the stand-in is not the
	// answer any more and the skip must not cover the object: a
	// Lord of Extinction with every graveyard empty is a real 0/0
	// and dies (#690).
	//
	// Set by applyOneEffectLocked, in the 7a and 7b buckets only,
	// after the effect's Apply has run — so it is true exactly when
	// an effect that defines P/T applied to this object. Layers 7c
	// (modify), 7d (counters) and 7e (switch) do not set it: they
	// move a number that something else has to have defined first.
	//
	// Not compared by sameCharacteristic: the CR 613.8 dependency
	// probe asks what an effect DOES to an object, and "an effect
	// defined the P/T" is a fact about the pass, not about the
	// object. The probe's applyRaw never writes it.
	//
	// Rebuilt from scratch on every recompute with the rest of the
	// Characteristic, so it never goes stale and the snapshot never
	// carries it.
	PTDefined bool

	// Restrictions is the S24 declaration-time restriction set:
	// "can't attack", "can't block", "can't be blocked", "its
	// activated abilities can't be activated". Like Controller it is
	// not a characteristic in the CR 109.3 sense — CR 613 gives
	// restrictions no layer at all, because a restriction is an
	// effect the SOURCE has rather than an ability the restricted
	// permanent has.
	//
	// It lives here anyway because the layer pass is the only thing
	// that knows which permanents a continuous effect applies to.
	// The field is only ever OR'd into and no layer clears it, so
	// the layer a restriction is written in and its timestamp are
	// both irrelevant — and "enchanted creature loses all abilities"
	// cannot strip a Pacifism, which is the rules-correct outcome.
	// restrictions.go has the taxonomy.
	Restrictions Restriction
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
// Colors come from printedColors: Card.Colors when the importer
// stamped it, the mana cost otherwise. (This comment used to say
// colours were not derived here at all — true at S16 sub-PR 1, wrong
// since S20 stamped Card.Colors, and wrong in a way that mattered:
// the cost-only derivation was what made every colour-indicator face
// read as colourless.)
func (c Card) printedCharacteristic() Characteristic {
	// CR 708.2, ADR 0069 decision 3: a face-down permanent IS a 2/2
	// creature with no name, text, subtypes, mana cost or colour.
	// That is not an effect applied to the real card — the object has
	// those characteristics — so it enters at layer 0, the baseline
	// the whole CR 613 pass is applied to. Every later layer and
	// every reader then sees the 2/2 for free.
	if c.FaceDownIsPermanent() {
		return faceDownCharacteristic(c)
	}
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
	// inline as plain data (S21 sub-PR 1). The catalog remains a
	// fallback and an override for cards that never go through deck
	// import — fixtures, tokens, and any spec that deliberately
	// states a keyword Scryfall doesn't.
	//
	// The two sources overlap for every catalog card that is also
	// imported from a decklist, so the merge (mergePrintedKeywords)
	// dedupes: a doubled "flash" is harmless to HasKeyword but renders
	// as two badges on the client's keyword row. A CUMULATIVE keyword
	// (prowess, toxic) is the one exception — see that function.
	//
	// The gate is the catalog KEY and not an oracle ID (ADR 0083
	// decision 3): a token has a key of its own since #521, so a
	// template that declares PrintedKeywords in its catalog entry is
	// read here like any card's. CatalogKey already answers "" for
	// an uncatalogued object.
	var catalogKeywords []string
	if key := CatalogKey(c); CatalogPrintedKeywords != nil && key != "" {
		catalogKeywords = CatalogPrintedKeywords(key)
	}
	abilities := mergePrintedKeywords(catalogKeywords, c.Keywords)
	return Characteristic{
		Power:      c.Power,
		Toughness:  c.Toughness,
		Types:      types,
		Subtypes:   subtypes,
		Supertypes: supertypes,
		Colors:     printedColors(c),
		Name:       c.Name,
		Abilities:  abilities,
		// CR 702.73a is a characteristic-defining ability, and
		// CR 613.2 applies CDAs before every other effect in their
		// layer — so the keyword→layer-4 projection belongs in the
		// baseline the layer pass starts from, and this is the ONE
		// place a printed changeling becomes the type fact (#670).
		// A layer-4 subtype SET clears it (Characteristic.SetSubtypes)
		// and a layer-4 grant re-sets it, in timestamp order.
		AllCreatureTypes: containsKeyword(abilities, KeywordChangeling),
		// The layer-0 baseline for control (CR 613.1b): who controls
		// this object absent any control-changing continuous effect.
		Controller: c.baseController(),
	}
}

// SetSubtypes is "is a Forest land" / "is an Elk creature" / "isn't a
// creature" — a layer-4 effect that REPLACES the subtype list rather
// than adding to it (CR 205.1b, CR 613.1d).
//
// Every layer-4 subtype replacement in the engine goes through here,
// because replacing the subtypes is also what takes "is every
// creature type" away: a creature that is set to be an Elk is an Elk
// and nothing else, whatever a Maskwood Nexus said earlier in the same
// layer (#670, and the second Maskwood ruling). An ADD —
// `c.Subtypes = append(c.Subtypes, "Swamp")` — deliberately does not
// go through here, because adding a type takes nothing away.
func (c *Characteristic) SetSubtypes(subtypes []string) {
	c.Subtypes = append([]string(nil), subtypes...)
	c.AllCreatureTypes = false
}

// clone returns a deep copy — the slices too, so a caller can mutate
// it without reaching back into the original. The layer engine's
// CR 613.8 dependency probe needs it (layer_dependency.go): a trial
// application must not leave a fingerprint on the real pass.
func (c Characteristic) clone() Characteristic {
	out := c
	out.Types = append([]string(nil), c.Types...)
	out.Subtypes = append([]string(nil), c.Subtypes...)
	out.Supertypes = append([]string(nil), c.Supertypes...)
	out.Colors = append([]string(nil), c.Colors...)
	out.Abilities = append([]string(nil), c.Abilities...)
	out.GrantedAbilities = append([]GrantedAbility(nil), c.GrantedAbilities...)
	return out
}

// baseController is the controller a permanent reverts to when every
// control-changing effect on it ends — the player who controlled it
// when it entered the battlefield, or its current controller for a
// card the layer engine has not seen enter (anything off the
// battlefield, every test fixture, the demo seed).
//
// Card.BaseController is captured lazily by the recompute and cleared
// on battlefield exit, so it is zero exactly when "the current
// controller IS the base" is true.
func (c Card) baseController() uuid.UUID {
	if c.BaseController != uuid.Nil {
		return c.BaseController
	}
	return c.Controller
}

// containsKeyword reports whether xs already holds kw. Used to read
// the changeling type fact off the merged printed abilities; the
// battlefield's Layer 6 keyword synth keeps its own copy over in
// cards/effects, where the name would otherwise collide with that
// package's test helpers.
func containsKeyword(xs []string, kw string) bool {
	for _, x := range xs {
		if x == kw {
			return true
		}
	}
	return false
}

// mergePrintedKeywords combines the catalog's own Spec.PrintedKeywords
// with the deck-imported Card.Keywords into the one printed-keyword
// baseline printedCharacteristic starts from. The two sources overlap
// for every catalog card that is also imported from a decklist (see
// the caller's comment), and how they combine depends on whether the
// keyword is CUMULATIVE (KeywordIsCumulative — prowess, toxic,
// CR 702.108b / 702.164b):
//
//   - a keyword that is NOT cumulative is a plain union: present in
//     either source, it appears once in the result, same as before
//     this function existed.
//   - a CUMULATIVE keyword takes the HIGHER of the two sources' own
//     repeat counts, never their sum. A catalog entry that declares
//     one "prowess" and an import that counts two off a doubled
//     oracle line ("Prowess, prowess", #1510) describe the same two
//     abilities, not three — summing would double what the import
//     already counted correctly, and a card that is ONLY imported (no
//     catalog entry) still gets exactly the import's own count.
//
// Catalog keywords are walked first so a card's own PrintedKeywords
// keep first-seen order ahead of the imported ones, matching the
// order the pre-#1510 loop produced for the common case of no
// repeats.
func mergePrintedKeywords(catalog, imported []string) []string {
	if len(catalog) == 0 && len(imported) == 0 {
		return nil
	}
	catalogCounts := countKeywords(catalog)
	importedCounts := countKeywords(imported)

	out := make([]string, 0, len(catalog)+len(imported))
	seen := map[string]bool{}
	add := func(kw string) {
		if seen[kw] {
			return
		}
		seen[kw] = true
		n := 1
		if KeywordIsCumulative(kw) {
			n = catalogCounts[kw]
			if importedCounts[kw] > n {
				n = importedCounts[kw]
			}
			if n < 1 {
				n = 1
			}
		}
		for i := 0; i < n; i++ {
			out = append(out, kw)
		}
	}
	for _, kw := range catalog {
		add(kw)
	}
	for _, kw := range imported {
		add(kw)
	}
	return out
}

// countKeywords tallies how many times each token appears in xs — the
// per-source repeat count mergePrintedKeywords compares to decide a
// cumulative keyword's merged count.
func countKeywords(xs []string) map[string]int {
	if len(xs) == 0 {
		return nil
	}
	counts := make(map[string]int, len(xs))
	for _, x := range xs {
		counts[x]++
	}
	return counts
}

// printedColors is the printed-colour rule for a whole card: the
// STAMPED Colors list when the importer gave us one, falling back to
// the colours in the mana cost.
//
// It used to be the cost alone, which is wrong for three classes of
// card and was a latent bug independent of faces:
//
//   - Devoid (CR 702.114) prints a cost with coloured pips and is
//     colourless anyway. Scryfall's `colors` says so; the cost does
//     not.
//   - CR 105.2c colour indicators — the coloured dot on a card with
//     no mana cost. Every transform back face is in this class:
//     Jace, Telepath Unbound has cost "" and reads as colourless
//     from the cost, blue from the indicator.
//   - Land faces of a modal DFC, which are correctly colourless but
//     used to be indistinguishable from "not stamped".
//
// Empty Colors still means "not stamped" rather than "colourless" —
// tokens and test fixtures never set it — so the cost fallback is
// kept rather than replaced. That matches the posture
// Card.EffectiveColors has taken since S20.
func printedColors(c Card) []string {
	if len(c.Colors) > 0 {
		return append([]string(nil), c.Colors...)
	}
	return printedColorsFromCost(c.ManaCost)
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
