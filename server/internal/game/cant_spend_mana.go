package game

// cant_spend_mana.go — #6227: "You can't spend mana to cast this
// spell" (Hogaak, Arisen Necropolis).
//
// CR 106.1 is what the clause takes away: "Players spend mana to pay
// costs, usually when casting spells and activating abilities." A
// spell that forbids it still has a total cost — CR 601.2f settles it
// the same way for Hogaak as for anything else — and still has to pay
// it in full, because CR 601.2h says "Partial payments are not
// allowed. Unpayable costs can't be paid." What is left to pay it with
// is everything that is NOT mana: convoke's taps (CR 702.51a, "you may
// tap an untapped creature … rather than pay that mana") and delve's
// exiles (CR 702.66a, "you may exile a card from your graveyard rather
// than pay that mana"). Both already subtract from the total in the one
// pricer (costAfterModifiersLocked), in that order, so this clause adds
// no arithmetic: it only has to make the REMAINDER unpayable.
//
// So the clause is not a cost component and not a cast ban. It is a
// restriction on the payment, which the engine already has one door
// for: ManaSpendContext (#352, #2556). The context carries NoMana, the
// pool's one admission gate refuses every token under it
// (allowsToken), and the auto-tapper makes no plan for it. Everything
// that has to agree with the payment falls out of that, because they
// all ask through the same two functions: the cast
// (applyCastCostLocked), its auto-tap (applyAutoTapLocked),
// internal/legal's affordability probe and the auto-tap preview. A
// cast whose convoke taps and delve exiles leave anything owed is
// refused for insufficient mana, with the symbols it could not pay;
// one that covers the whole of it pays with no mana at all.
//
// Two things the clause deliberately does NOT touch. Mana abilities
// still work (CR 601.2g's window is about the mana the payment may use,
// not about adding it), so a player who taps a Forest and then casts
// Hogaak keeps the {G} floating. And a non-mana cost beside the mana —
// an additional cost, a Phyrexian symbol paid with life (CR 107.4f) —
// is not mana and is paid as it always was.

// CatalogCantSpendMana is the catalog hook the effects package wires at
// init through CardDef.CantSpendMana. Nil, or false, means the card
// says nothing about spending mana — which is every card but one.
var CatalogCantSpendMana func(key string) bool

// CantSpendManaFor reports whether the card being cast says "You can't
// spend mana to cast this spell", read off the catalog entry of the
// face being cast, exactly as DelveFor reads delve. A face-down cast
// (CR 708.4) has no catalog key and no text, so it never says it.
func CantSpendManaFor(card Card) bool {
	if CatalogCantSpendMana == nil {
		return false
	}
	key := CatalogKey(card)
	if key == "" {
		return false
	}
	return CatalogCantSpendMana(key)
}
