package game

import (
	"fmt"
	"os"
	"sync/atomic"
	"testing"
)

// printed_cache.go memoises a card's layer-0 baseline — the
// Characteristic printedCharacteristic builds from its printed fields
// — so the view stops re-deriving it on every read (#1498).
//
// Off the battlefield there is no layer pass and Card.effective is
// nil, so every Effective() call used to rebuild the baseline from
// scratch: a type-line parse, a colour slice, the catalog/import
// keyword merge. The view asks several times per card (the card view
// itself, CurrentPower, PowerForComparison, CurrentToughness) for
// every card in every hand, graveyard, exile and command zone, on
// every frame. That was ~8% of a random bot table's CPU, ParseTypeLine
// alone 5%.
//
// # Why it cannot go stale
//
// The cache is not invalidated anywhere. It does not need to be,
// because it is never trusted blind: an entry records every input the
// baseline is built from (printedInputs), and each read compares the
// card's live inputs against the recorded ones and rebuilds on any
// difference. The baseline is a pure function of those inputs —
// printedInputs.characteristic reads nothing else — so "same inputs"
// is "same answer" by construction, whatever changed the card in
// between: a face turned (transform, MDFC, flip, split, adventure,
// omen, day/night), a copy effect rewriting the printed values (ADR
// 0043), a deck import, a snapshot restore, an undo, or a test that
// writes TypeLine in place. A changed card is a cache MISS, and a miss
// is today's uncached build — slower, never wrong.
//
// The inputs include the catalog's printed-keyword answer for the
// card's key, re-asked on every read, so a test that swaps
// CatalogPrintedKeywords, or registers a spec after a card was
// stamped, is covered too. The recorded slices are the entry's own
// copies, so writing an element of Card.Colors in place cannot make a
// stale entry compare equal to itself.
//
// Two things are deliberately NOT cached:
//
//   - A face-down permanent (CR 708.2). Its baseline is the 2/2 body
//     (faceDownCharacteristic), which reads FaceDownKind and
//     FaceDownListed rather than the printed fields. It is rare, and
//     keeping it out keeps the key small.
//   - The controller. It is per-instance state, not printed data,
//     and changes far more often than the rest; it is stamped onto the
//     copy every read hands out (baseController), as before.
//
// # Who writes it, and why that is race-free
//
// An entry is immutable once built. Card.printed is replaced, never
// written through, so a clone sharing an entry with its source
// (cloneCard copies the pointer, exactly as it copies `effective`) is
// as safe as two cards sharing a string: neither can change what the
// other reads, and each validates against its own fields.
//
// The pointer is written only by stampPrinted, from the zone
// insertions (Zone.PushTop, PushBottom, InsertFromTop) and the
// snapshot restore — all under the game's WRITE lock, like every other
// zone mutation. Readers never write it: the view runs under the READ
// lock, concurrently with the bot enumerators, and a reader that
// filled the cache would be a data race. A reader that finds no entry,
// or a stale one, builds the answer and throws it away.
//
// # Read-only slices
//
// A cache hit hands out the entry's own Types / Subtypes / Supertypes
// / Colors / Abilities slices — the contract Effective() has always
// had for a battlefield card, whose answer shares the layer cache's
// slices. Every slice is clipped to its length, so an append (the
// keyword-counter fold in Effective, a caller's) always reallocates
// rather than writing past the end into the entry. Writing an ELEMENT
// in place would corrupt the entry; nothing in the tree does, and the
// test-binary check below fails the first read after one.
//
// The layer engine does NOT read the cache: its pass writes the
// baseline in place (a layer-6 "loses flying" filters Abilities
// in place), so it keeps building a fresh, owned one through
// printedCharacteristic.

// printedInputs is everything printedCharacteristic reads off a
// face-up card, and nothing else. printedInputs.characteristic is a
// pure function of it.
type printedInputs struct {
	typeLine string
	name     string
	manaCost string

	power     int
	toughness int

	colors          []string // Card.Colors (empty: derive — none if devoid, else the cost's)
	keywords        []string // Card.Keywords, the import's keywords
	catalogKeywords []string // CatalogPrintedKeywords(CatalogKey(c))
}

// printedInputsOf reads the inputs off a face-up card. The slices
// alias the card's (and the catalog's); only stampPrinted copies them.
func printedInputsOf(c *Card) printedInputs {
	in := printedInputs{
		typeLine:  c.TypeLine,
		name:      c.Name,
		manaCost:  c.ManaCost,
		power:     c.Power,
		toughness: c.Toughness,
		colors:    c.Colors,
		keywords:  c.Keywords,
	}
	// The gate is the catalog KEY and not an oracle ID (ADR 0083
	// decision 3): a token has a key of its own since #521, so a
	// template that declares PrintedKeywords in its catalog entry is
	// read here like any card's. catalogKeyOf already answers "" for
	// an uncatalogued object.
	if CatalogPrintedKeywords != nil {
		if key := catalogKeyOf(c); key != "" {
			in.catalogKeywords = CatalogPrintedKeywords(key)
		}
	}
	return in
}

// characteristic builds the layer-0 baseline from the inputs. The
// controller is left zero; the caller stamps it.
func (in *printedInputs) characteristic() Characteristic {
	supertypes, types, subtypes := ParseTypeLine(in.typeLine)
	return in.characteristicFrom(supertypes, types, subtypes)
}

// characteristicFrom is characteristic with the type line already
// parsed — split out so the test-binary check can rebuild without
// counting as a ParseTypeLine call (parseTypeLineCalls).
func (in *printedInputs) characteristicFrom(supertypes, types, subtypes []string) Characteristic {
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
	abilities := mergePrintedKeywords(in.catalogKeywords, in.keywords)
	return Characteristic{
		Power:      in.power,
		Toughness:  in.toughness,
		Types:      types,
		Subtypes:   subtypes,
		Supertypes: supertypes,
		// CR 702.114a (#2152): devoid is read off the same merged
		// list as changeling below, so the import's keyword and the
		// catalog's both count, and the entry's key (keywords,
		// catalogKeywords) already records the input it depends on.
		Colors:    printedColorsOf(in.colors, in.manaCost, containsKeyword(abilities, KeywordDevoid)),
		Name:      in.name,
		Abilities: abilities,
		// CR 702.73a is a characteristic-defining ability, and
		// CR 613.2 applies CDAs before every other effect in their
		// layer — so the keyword→layer-4 projection belongs in the
		// baseline the layer pass starts from, and this is the ONE
		// place a printed changeling becomes the type fact (#670).
		// A layer-4 subtype SET clears it (Characteristic.SetSubtypes)
		// and a layer-4 grant re-sets it, in timestamp order.
		AllCreatureTypes: containsKeyword(abilities, KeywordChangeling),
	}
}

// sameAs reports whether two input sets would build the same
// baseline. Slices compare by content, and nil equals empty because
// characteristic treats them alike (len checks throughout).
func (in *printedInputs) sameAs(o *printedInputs) bool {
	return in.typeLine == o.typeLine &&
		in.name == o.name &&
		in.manaCost == o.manaCost &&
		in.power == o.power &&
		in.toughness == o.toughness &&
		sameStrings(in.colors, o.colors) &&
		sameStrings(in.keywords, o.keywords) &&
		sameStrings(in.catalogKeywords, o.catalogKeywords)
}

func sameStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// printedEntry is one memoised baseline: the inputs it was built from
// (the entry's own copies) and the answer, Controller zero, every
// slice clipped to its length. Immutable once built.
type printedEntry struct {
	in printedInputs
	ch Characteristic
}

// newPrintedEntry builds an entry for the given inputs.
func newPrintedEntry(in printedInputs) *printedEntry {
	ch := in.characteristic()
	in.colors = ownStrings(in.colors)
	in.keywords = ownStrings(in.keywords)
	in.catalogKeywords = ownStrings(in.catalogKeywords)
	ch.Types = clipStrings(ch.Types)
	ch.Subtypes = clipStrings(ch.Subtypes)
	ch.Supertypes = clipStrings(ch.Supertypes)
	ch.Colors = clipStrings(ch.Colors)
	ch.Abilities = clipStrings(ch.Abilities)
	return &printedEntry{in: in, ch: ch}
}

func ownStrings(s []string) []string {
	if len(s) == 0 {
		return nil
	}
	return append([]string(nil), s...)
}

func clipStrings(s []string) []string { return s[:len(s):len(s)] }

// stampPrinted (re)builds the card's cached baseline when it has none
// or the one it has no longer matches its printed values. A face-down
// permanent drops its entry: its baseline is not cached.
//
// Caller must hold the game's WRITE lock, or own the card outright (a
// zone being built that no other goroutine can see yet). Never call
// this from a read path — see the file comment.
func (c *Card) stampPrinted() {
	if c.faceDownPermanent() {
		c.printed = nil
		return
	}
	in := printedInputsOf(c)
	if e := c.printed; e != nil && e.in.sameAs(&in) {
		return
	}
	c.printed = newPrintedEntry(in)
}

// printedShared is the layer-0 baseline for read-only use: the cached
// entry when its inputs still match the card's, a fresh build when
// they do not. The slices may be the entry's own — callers must not
// write their elements (see "Read-only slices" above). The layer
// engine, which does, uses printedCharacteristic instead.
func printedShared(c *Card) Characteristic {
	// CR 708.2, ADR 0069 decision 3: a face-down permanent IS a 2/2
	// creature with no name, text, subtypes, mana cost or colour. Not
	// cached; see the file comment.
	if c.faceDownPermanent() {
		return faceDownCharacteristic(*c)
	}
	in := printedInputsOf(c)
	var ch Characteristic
	if e := c.printed; e != nil && e.in.sameAs(&in) {
		ch = e.ch
		if printedCacheChecks {
			checkPrintedEntry(c, e, &in)
		}
	} else {
		ch = in.characteristic()
	}
	// The layer-0 baseline for control (CR 613.1b): who controls this
	// object absent any control-changing continuous effect.
	ch.Controller = c.baseController()
	return ch
}

// printedTypeParts is the card's parsed PRINTED type line, read-only:
// the cached entry's slices when it was built from this exact type
// line, a fresh ParseTypeLine otherwise. The type parts depend on the
// type line alone, so only that input has to match. It ignores face-
// down state — callers that care (HasSubtype, HasSupertype) branch on
// FaceDownIsPermanent first, as they did around ParseTypeLine.
func printedTypeParts(c *Card) (supertypes, types, subtypes []string) {
	if e := c.printed; e != nil && e.in.typeLine == c.TypeLine {
		if printedCacheChecks {
			checkPrintedTypeParts(c, e)
		}
		return e.ch.Supertypes, e.ch.Types, e.ch.Subtypes
	}
	return ParseTypeLine(c.TypeLine)
}

// PrintedTypeLineIs reports whether the card's printed type line
// parses to exactly these supertypes, types and subtypes — the
// protocol layer's "has any layer-4 effect changed this card's types"
// test (effectiveTypeLine), answered off the cached parse rather than
// a fresh ParseTypeLine per card per view (#1498).
func (c *Card) PrintedTypeLineIs(supertypes, types, subtypes []string) bool {
	ps, pt, pu := printedTypeParts(c)
	return sameStrings(ps, supertypes) && sameStrings(pt, types) && sameStrings(pu, subtypes)
}

// printedCacheChecks turns every cache hit into a hit-and-verify: the
// baseline is rebuilt from the card's live fields and compared with
// the entry, and a difference panics. On in every test binary, so the
// whole suite — every face change, copy effect, restore and undo it
// exercises — doubles as the cache's invalidation test; off in
// production, where the rebuild is exactly the cost being saved.
//
// A mismatch can only mean the baseline read an input printedInputs
// does not record, or that a reader wrote an element of a slice it
// was handed (see "Read-only slices").
//
// CMDCTRL_PRINTED_CACHE_CHECKS=0 turns the checks off in a test binary,
// for profiling one (the bot soak): with them on, every hit pays the
// rebuild the cache exists to save, and the profile says so. It cannot
// turn them on in production.
var printedCacheChecks = testing.Testing() && os.Getenv("CMDCTRL_PRINTED_CACHE_CHECKS") != "0"

func checkPrintedEntry(c *Card, e *printedEntry, in *printedInputs) {
	fresh := in.characteristicFrom(parseTypeLine(in.typeLine))
	if !samePrintedCharacteristic(fresh, e.ch) {
		panic(fmt.Sprintf("game: stale printed-characteristic cache for %q (%s): cached %+v, rebuilt %+v",
			c.Name, c.InstanceID, e.ch, fresh))
	}
}

func checkPrintedTypeParts(c *Card, e *printedEntry) {
	s, t, u := parseTypeLine(c.TypeLine)
	if !sameStrings(s, e.ch.Supertypes) || !sameStrings(t, e.ch.Types) || !sameStrings(u, e.ch.Subtypes) {
		panic(fmt.Sprintf("game: stale printed type-line cache for %q (%s): cached %v %v %v, parsed %v %v %v",
			c.Name, c.InstanceID, e.ch.Supertypes, e.ch.Types, e.ch.Subtypes, s, t, u))
	}
}

// samePrintedCharacteristic compares every field
// printedInputs.characteristic sets.
// TestPrintedCharacteristicSetsOnlyTheComparedFields fails if the
// builder starts setting one this does not compare.
func samePrintedCharacteristic(a, b Characteristic) bool {
	return a.Power == b.Power &&
		a.Toughness == b.Toughness &&
		a.Name == b.Name &&
		a.AllCreatureTypes == b.AllCreatureTypes &&
		a.Controller == b.Controller &&
		sameStrings(a.Types, b.Types) &&
		sameStrings(a.Subtypes, b.Subtypes) &&
		sameStrings(a.Supertypes, b.Supertypes) &&
		sameStrings(a.Colors, b.Colors) &&
		sameStrings(a.Abilities, b.Abilities)
}

// parseTypeLineCalls counts ParseTypeLine calls in a test binary, for
// the view's steady-state cost test. Not counted in production, where
// countParseTypeLine is false and the counter is one predictable branch.
var (
	countParseTypeLine = testing.Testing()
	parseTypeLineCalls atomic.Int64
)

// ParseTypeLineCallsForTest returns how many times ParseTypeLine has
// run in this test binary. Always zero in production.
func ParseTypeLineCallsForTest() int64 { return parseTypeLineCalls.Load() }
