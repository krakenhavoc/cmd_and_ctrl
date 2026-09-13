package effects

// Brokers Hideout — Land (EDHREC rank 1304):
//
//	"When this land enters, sacrifice it. When you do, search your
//	 library for a basic Forest, Plains, or Island card, put it onto
//	 the battlefield tapped, then shuffle and you gain 1 life."
//
// The Bant member of the Streets of New Capenna "Overlook" cycle
// Riveteers Overlook established (batch 06). Same shape through the
// b08OverlookLand helper, same declared gap: the printed reflexive
// trigger is folded into the entry trigger, so the sacrifice and the
// search are one stack item.
func init() {
	Register(b08OverlookLand("bd002797-a545-4bee-88bf-b878436e7cca", "Brokers Hideout", "Forest", "Plains", "Island"))
}
