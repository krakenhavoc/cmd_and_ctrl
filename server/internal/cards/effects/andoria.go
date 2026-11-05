package effects

// Andoria — Land — Planet (Star Trek, #4973):
//
//	"This land enters tapped.
//	 When this land enters, you gain 1 life.
//	 {T}: Add {U} or {R}."
//
// A gain land (gain_lands.go) with the Planet subtype, which the
// printed type line carries for anything that cares. Registered
// through the cycle's shared body, registerGainLand.
//
// No simplification.
func init() {
	registerGainLand("79134f5e-94db-4a40-ba75-2569b839cf35", "Andoria", "U", "R")
}
