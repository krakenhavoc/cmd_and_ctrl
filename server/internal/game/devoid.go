package game

// devoid.go — CR 702.114, devoid (#2152).
//
//	702.114a Devoid is a characteristic-defining ability. "Devoid"
//	means "This object is colorless." This ability functions
//	everywhere, even outside the game. See rule 604.3.
//
// # The bug it closes
//
// Scryfall already computes a devoid card's colours as an empty list,
// and the deck importer stamps that list onto Card.Colors. But an
// empty Card.Colors has always meant "not stamped" to the engine —
// tokens and test fixtures never set it — and an unstamped card takes
// its colours from its mana cost. So every devoid card was the colour
// of its pips: Ugin's Binding was blue, a "nonblue" target clause
// refused it, and a "colorless spell" payoff did not see it.
//
// # Why a keyword, and not "empty means empty"
//
// Making the importer's empty list mean colourless (a nil-versus-empty
// distinction threaded through import, snapshot and clone) would fix
// imported cards only, and quietly: a fixture, a token template or a
// catalog-only card with a coloured cost and no stamped colours would
// still read as coloured, and nothing would say why a JSON round trip
// that turned [] into null changed a card's colour. Devoid is printed
// as a keyword, Scryfall lists it in `keywords`, and the engine already
// carries printed keywords on two roads — Card.Keywords from the
// importer, Spec.PrintedKeywords from the catalog. So devoid rides
// those, exactly as changeling (the other CDA keyword, CR 702.73a)
// does, and the colour rule reads it.
//
// # Where it is read
//
// In the colour DERIVATION and nowhere else (printedColorsOf): a card
// whose Colors is empty and that prints devoid is colourless rather
// than the colour of its mana cost. That is layer 0, the baseline the
// CR 613 pass starts from, which is where a CDA belongs (CR 613.3:
// CDAs apply first within their layer), so:
//
//   - it holds in every zone — hand, library, stack, battlefield,
//     graveyard, exile, command — because every colour read goes
//     through that baseline (Effective, EffectiveColors, HasColor,
//     IsColorless, SourceCharacteristics);
//   - a layer-5 colour effect still applies on top of it: Cerulean
//     Wisps makes a devoid creature blue until end of turn;
//   - a layer-6 "loses all abilities" does not make it coloured
//     again, because layer 5 has already been applied (CR 613.1);
//   - a copy is colourless too, because Keywords and the oracle ID
//     are copiable values (CR 707.2) and they carry the ability.
//
// A NON-EMPTY Colors wins over devoid, deliberately. Every source of a
// non-empty list already accounts for the card's colour-defining
// abilities: Scryfall's computed colours (always empty for a devoid
// card), a colour indicator, or a copy effect's "except it's black"
// (eternalize, The Scarab God), where CR 707.9d says the copied
// object's colour-defining abilities are NOT copied. Reading devoid
// over a stamped colour would make an eternalized Eldrazi colourless
// instead of the black Zombie it is.
//
// # What it does not touch
//
// Colour IDENTITY (CR 903.4). Identity comes from the mana symbols in
// the cost and rules text, and devoid does not remove them: an
// imported card carries Scryfall's color_identity verbatim, and an
// unstamped one (printedIdentityOf) falls back to its mana cost when
// its colours come up empty — which for a devoid card they now do.

// KeywordDevoid is the canonical token for devoid (CR 702.114a).
// Named for the reason KeywordChangeling is: it is read from more than
// one package, and a typo would silently give every Eldrazi its pips'
// colours back.
const KeywordDevoid = "devoid"

// printsDevoid reports whether the card's own printed keywords say it
// is devoid: the importer's Card.Keywords or the catalog's
// Spec.PrintedKeywords — the two lists printedInputs merges into the
// baseline's abilities, so this and the cached baseline cannot
// disagree. The catalog is asked only when the card's own list does not
// already answer, and only for a catalogued object.
//
// Printed keywords only: devoid is a CDA, and CR 604.3a makes a CDA one
// that is printed on the card (or set by a copy's copiable values, which
// write Keywords). A keyword granted by an effect is not one.
func (c *Card) printsDevoid() bool {
	if containsKeyword(c.Keywords, KeywordDevoid) {
		return true
	}
	if CatalogPrintedKeywords == nil {
		return false
	}
	key := catalogKeyOf(c)
	return key != "" && containsKeyword(CatalogPrintedKeywords(key), KeywordDevoid)
}
