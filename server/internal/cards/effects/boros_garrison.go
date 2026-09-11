package effects

// Boros Garrison — Land (EDHREC rank 494):
//
//	"This land enters tapped.
//	 When this land enters, return a land you control to its owner's
//	 hand.
//	 {T}: Add {R}{W}."
//
// The bounce-land shape — see b04BounceLand in batch04_helpers.go
// for the machinery and the one modelling note (the bounce is a
// choice modelled as a target). No other simplification.
func init() {
	Register(b04BounceLand("8fa3ac81-3dfe-4565-be99-5554f7597b4b", "Boros Garrison", "{R}{W}"))
}
