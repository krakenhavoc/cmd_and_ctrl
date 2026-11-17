package effects

// Sand Silos — Land (Fallen Empires, #5637):
//
//	"This land enters tapped.
//	 You may choose not to untap this land during your untap step.
//	 At the beginning of your upkeep, if this land is tapped, put a
//	 storage counter on it.
//	 {T}, Remove any number of storage counters from this land: Add {U}
//	 for each storage counter removed this way."
//
// The blue storage land: Bottomless Vault's rows in {U}, built by
// storageLand (storage_lands.go). The upkeep trigger is an intervening
// "if" (CR 603.4), read again as it resolves, and the payout counts the
// counters that actually came off.
//
// No simplification.
func init() {
	Register(storageLand("48a830f1-8965-4f97-b3d8-ca98eab1ba33", "Sand Silos", "U"))
}
