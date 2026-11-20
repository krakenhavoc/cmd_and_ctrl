package effects

// Molten Slagheap — Land (Time Spiral, #5596):
//
//	"{T}: Add {C}.
//	 {1}, {T}: Put a storage counter on this land.
//	 {1}, Remove X storage counters from this land: Add X mana in any
//	 combination of {B} and/or {R}."
//
// The black-red storage land, built by tspB2StorageLand: the counter is a
// plain activated ability that uses the stack, and the payout is a mana
// ability that pays {1}, removes the X counters announced and adds X
// mana, each {B} or {R} (CR 605.1a).
//
// No simplification.
func init() {
	Register(tspB2StorageLand("33587cb2-0fd3-4e4c-bc5e-e7299cc9dab5", "Molten Slagheap", "B", "R"))
}
