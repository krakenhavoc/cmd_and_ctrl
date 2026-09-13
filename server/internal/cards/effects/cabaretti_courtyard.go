package effects

// Cabaretti Courtyard — Land (EDHREC rank 968):
//
//	"When this land enters, sacrifice it. When you do, search your
//	 library for a basic Mountain, Forest, or Plains card, put it onto
//	 the battlefield tapped, then shuffle and you gain 1 life."
//
// The Naya member of the Streets of New Capenna "Overlook" cycle
// Riveteers Overlook established (batch 06). Same shape through the
// b08OverlookLand helper, same declared gap: the printed reflexive
// trigger is folded into the entry trigger, so the sacrifice and the
// search are one stack item.
func init() {
	Register(b08OverlookLand("65424bea-fd53-4f85-9757-0b91a6d40ba4", "Cabaretti Courtyard", "Mountain", "Forest", "Plains"))
}
