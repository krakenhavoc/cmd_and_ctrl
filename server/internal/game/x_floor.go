package game

// x_floor.go — a printed floor on a spell's announced X (#6221).
//
//	Ertai's Meddling  X can't be 0.
//
// CR 107.3a makes X a number the caster chooses and announces (CR 601.2b),
// and a printed "X can't be 0" bounds that announcement from below, as
// XCeiling does from above. The cast is refused below the floor, in the same
// place the ceiling is checked (castSpellLocked). The legal-move enumerator
// needs nothing more for the one card that prints it: an X spell that
// declares Spec.XMatters is never offered at X=0.

// CatalogXFloor is the catalog hook for a card's printed X floor;
// carddef.go's init reads it off the CardDef. Nil (no catalog) means no
// card declares one.
var CatalogXFloor func(oracleID string) int

// XFloorFor is the printed X floor the catalog declares for `key`, or zero
// for the overwhelming majority of cards, which print none.
func XFloorFor(key string) int {
	if CatalogXFloor == nil || key == "" {
		return 0
	}
	return CatalogXFloor(key)
}
