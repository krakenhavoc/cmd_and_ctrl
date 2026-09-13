package effects

// Obscura Storefront — Land (EDHREC rank 1645):
//
//	"When this land enters, sacrifice it. When you do, search your
//	 library for a basic Plains, Island, or Swamp card, put it onto
//	 the battlefield tapped, then shuffle and you gain 1 life."
//
// The Esper member of the Streets of New Capenna "Overlook" cycle
// Riveteers Overlook established (batch 06). Same shape through the
// b08OverlookLand helper, same declared gap: the printed reflexive
// trigger is folded into the entry trigger, so the sacrifice and the
// search are one stack item.
func init() {
	Register(b08OverlookLand("dc31a6f8-6228-4a25-b937-5d8d78514333", "Obscura Storefront", "Plains", "Island", "Swamp"))
}
