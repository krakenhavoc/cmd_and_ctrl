package effects

// Saltcrusted Steppe — Land (Time Spiral, #5596):
//
//	"{T}: Add {C}.
//	 {1}, {T}: Put a storage counter on this land.
//	 {1}, Remove X storage counters from this land: Add X mana in any
//	 combination of {G} and/or {W}."
//
// The green-white storage land, built by tspB2StorageLand: the counter is a
// plain activated ability that uses the stack, and the payout is a mana
// ability that pays {1}, removes the X counters announced and adds X
// mana, each {G} or {W} (CR 605.1a).
//
// No simplification.
func init() {
	Register(tspB2StorageLand("021e4165-2f02-4bd4-86ca-cb7bf4c9e23d", "Saltcrusted Steppe", "G", "W"))
}
