package effects

// Bajor — Land — Planet (Star Trek, #4973):
//
//	"This land enters tapped.
//	 When this land enters, you gain 1 life.
//	 {T}: Add {R} or {W}."
//
// A gain land (gain_lands.go) with the Planet subtype, which the
// printed type line carries for anything that cares. Registered
// through the cycle's shared body, registerGainLand.
//
// No simplification.
func init() {
	registerGainLand("93d7f6e7-b6b0-48f3-891a-73970826af3b", "Bajor", "R", "W")
}
