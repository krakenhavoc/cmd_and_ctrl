package effects

// Selesnya Sanctuary — Land (EDHREC rank 572):
//
//	"This land enters tapped.
//	 When this land enters, return a land you control to its owner's
//	 hand.
//	 {T}: Add {G}{W}."
//
// The bounce-land shape — see b04BounceLand in batch04_helpers.go.
func init() {
	Register(b04BounceLand("00ef1c55-dea1-4564-bd57-66de86cba4df", "Selesnya Sanctuary", "{G}{W}"))
}
