package effects

// Fungal Reaches — Land (Time Spiral, #5596):
//
//	"{T}: Add {C}.
//	 {1}, {T}: Put a storage counter on this land.
//	 {1}, Remove X storage counters from this land: Add X mana in any
//	 combination of {R} and/or {G}."
//
// The red-green storage land, built by tspB2StorageLand: the counter is a
// plain activated ability that uses the stack, and the payout is a mana
// ability that pays {1}, removes the X counters announced and adds X
// mana, each {R} or {G} (CR 605.1a).
//
// No simplification.
func init() {
	Register(tspB2StorageLand("6f18ea44-3efa-4a45-abc6-86a0627e40f2", "Fungal Reaches", "R", "G"))
}
