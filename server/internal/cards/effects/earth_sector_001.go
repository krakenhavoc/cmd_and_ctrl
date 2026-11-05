package effects

// Earth, Sector 001 — Land — Planet (Star Trek, #4973):
//
//	"This land enters tapped.
//	 When this land enters, you gain 1 life.
//	 {T}: Add {G} or {U}."
//
// A gain land (gain_lands.go) with the Planet subtype, which the
// printed type line carries for anything that cares. Registered
// through the cycle's shared body, registerGainLand.
//
// No simplification.
func init() {
	registerGainLand("c75d87b0-7de9-4c50-b781-9749d67c1e07", "Earth, Sector 001", "G", "U")
}
