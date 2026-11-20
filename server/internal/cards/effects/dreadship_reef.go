package effects

// Dreadship Reef — Land (Time Spiral, #5596):
//
//	"{T}: Add {C}.
//	 {1}, {T}: Put a storage counter on this land.
//	 {1}, Remove X storage counters from this land: Add X mana in any
//	 combination of {U} and/or {B}."
//
// The blue-black storage land, built by tspB2StorageLand: the counter is a
// plain activated ability that uses the stack, and the payout is a mana
// ability that pays {1}, removes the X counters announced and adds X
// mana, each {U} or {B} (CR 605.1a).
//
// No simplification.
func init() {
	Register(tspB2StorageLand("130a8cf5-1354-4d17-91c8-c073642eb3db", "Dreadship Reef", "U", "B"))
}
