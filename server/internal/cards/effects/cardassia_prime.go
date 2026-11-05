package effects

// Cardassia Prime — Land — Planet (Star Trek, #4973):
//
//	"This land enters tapped.
//	 When this land enters, you gain 1 life.
//	 {T}: Add {W} or {B}."
//
// A gain land (gain_lands.go) with the Planet subtype, which the
// printed type line carries for anything that cares. Registered
// through the cycle's shared body, registerGainLand.
//
// No simplification.
func init() {
	registerGainLand("aff720b4-b8ef-46ea-9a3d-e101999e6526", "Cardassia Prime", "W", "B")
}
