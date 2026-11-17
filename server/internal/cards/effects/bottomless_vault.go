package effects

// Bottomless Vault — Land:
//
//	"This land enters tapped.
//	 You may choose not to untap this land during your untap step.
//	 At the beginning of your upkeep, if this land is tapped, put a
//	 storage counter on it.
//	 {T}, Remove any number of storage counters from this land: Add {B}
//	 for each storage counter removed this way."
//
// The black storage land, built by storageLand (storage_lands.go),
// which Fallen Empires' four storage lands share (#5637). The tapped
// entry is the self-replacement; the untap opt-out keeps it tapped
// through the untap step so the upkeep trigger, an intervening "if"
// (CR 603.4), has something to count. The payout counts the counters
// that actually came off, so removing none adds nothing.
//
// No simplification.
func init() {
	Register(storageLand("e43413e4-be17-49af-978a-26210d05f52a", "Bottomless Vault", "B"))
}
