package effects

// Ferenginar — Land — Planet (Star Trek, #4973):
//
//	"This land enters tapped.
//	 When this land enters, you gain 1 life.
//	 {T}: Add {B} or {G}."
//
// A gain land (gain_lands.go) with the Planet subtype, which the
// printed type line carries for anything that cares. Registered
// through the cycle's shared body, registerGainLand.
//
// No simplification.
func init() {
	registerGainLand("2e0753ad-7f76-4d26-8e87-6f201378f333", "Ferenginar", "B", "G")
}
