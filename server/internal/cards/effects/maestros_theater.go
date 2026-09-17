package effects

// Maestros Theater — Land (EDHREC rank 1270):
//
//	"When this land enters, sacrifice it. When you do, search your
//	 library for a basic Island, Swamp, or Mountain card, put it onto
//	 the battlefield tapped, then shuffle and you gain 1 life."
//
// The Grixis member of the Streets of New Capenna "Overlook" cycle
// Riveteers Overlook established (batch 06) and Cabaretti Courtyard
// joined (batch 08). Same shape through the b08OverlookLand helper,
// and since #636 the same two stack items: the entry trigger
// sacrifices the land and a CR 603.12 reflexive trigger does the
// searching, with a response window between them.
func init() {
	Register(b08OverlookLand("9464ddf2-4bcb-44f6-b945-89a132544de6", "Maestros Theater", "Island", "Swamp", "Mountain"))
}
