package effects

// Dwarven Hold — Land (Fallen Empires, #5637):
//
//	"This land enters tapped.
//	 You may choose not to untap this land during your untap step.
//	 At the beginning of your upkeep, if this land is tapped, put a
//	 storage counter on it.
//	 {T}, Remove any number of storage counters from this land: Add {R}
//	 for each storage counter removed this way."
//
// The red storage land: Bottomless Vault's rows in {R}, built by
// storageLand (storage_lands.go). The upkeep trigger is an intervening
// "if" (CR 603.4), read again as it resolves, and the payout counts the
// counters that actually came off.
//
// No simplification.
func init() {
	Register(storageLand("4a6625bd-3dd2-45f1-8dc9-034c833fa90c", "Dwarven Hold", "R"))
}
