package effects

// Icatian Store — Land (Fallen Empires, #5637):
//
//	"This land enters tapped.
//	 You may choose not to untap this land during your untap step.
//	 At the beginning of your upkeep, if this land is tapped, put a
//	 storage counter on it.
//	 {T}, Remove any number of storage counters from this land: Add {W}
//	 for each storage counter removed this way."
//
// The white storage land: Bottomless Vault's rows in {W}, built by
// storageLand (storage_lands.go). The upkeep trigger is an intervening
// "if" (CR 603.4), read again as it resolves, and the payout counts the
// counters that actually came off.
//
// No simplification.
func init() {
	Register(storageLand("87a0e0b9-6c2d-47a4-a3ed-7e0ae62fbffc", "Icatian Store", "W"))
}
