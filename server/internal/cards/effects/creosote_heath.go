package effects

// Creosote Heath — Land — Desert (EDHREC rank 4012):
//
//	"This land enters tapped.
//	 When this land enters, it deals 1 damage to target opponent.
//	 {T}: Add {G} or {W}."
//
// The Selesnya member of Outlaws of Thunder Junction's tapped Desert
// duals — Abraded Bluffs, Bristling Backwoods and Jagged Barrens are
// already in the catalog, and this is the same row of the same table,
// built by b36DesertDual.
//
// It is in the batch for the reason the other rows were: a cycle is
// only worth a shared builder if the builder is actually shared, and
// each new row is a free check that the colour pair and the damage
// trigger did not get transposed.
//
// The Desert subtype matters to the cards that count Deserts
// (Hazezon, Ramunap Ruins); it carries no intrinsic ability of its
// own, so nothing else needs declaring for it.
//
// No simplification.
func init() {
	Register(b36DesertDual("c116b787-5f7e-47ef-a694-58709770dd32", "Creosote Heath", "G", "W"))
}
