package effects

// Rakdos Carnarium — Land (EDHREC rank 504):
//
//	"This land enters tapped.
//	 When this land enters, return a land you control to its owner's
//	 hand.
//	 {T}: Add {B}{R}."
//
// The bounce-land shape — see b04BounceLand in batch04_helpers.go.
func init() {
	Register(b04BounceLand("0a023964-2905-4928-9c3e-dc63e6ebd218", "Rakdos Carnarium", "{B}{R}"))
}
