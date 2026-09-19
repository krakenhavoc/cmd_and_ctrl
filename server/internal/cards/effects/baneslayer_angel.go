package effects

// Baneslayer Angel — "Flying, first strike, lifelink, protection from
// Demons and from Dragons."
//
// 5/5 for 3WW, and the catalog's proof that a protection quality can
// be a SUBTYPE rather than a colour (#662, CR 702.16b-f). Two
// abilities, not one (CR 702.16m): "protection from Demons and from
// Dragons" is two tokens, because each quality is checked on its own
// and a Demon Dragon is refused by either.
//
// A CHANGELING SOURCE COUNTS AS BOTH. CR 702.73a makes a changeling
// every creature type, and #939 put that in the layer-4 flag
// Characteristic.AllCreatureTypes rather than in ~345 subtypes — so
// game.ProtectedFrom asks the flag as well as the subtype list, and a
// Mistform Ultimus really cannot block her, enchant her or damage
// her. That fall-out is the reason the quality reader is one function
// and not a colour check per consumer.
//
// Multi-keyword card — the PrintedKeywords slot feeds all five
// as strings into Characteristic.Abilities; combat reads the three
// combat keywords via HasKeyword and the two protections via
// game.ProtectionQualities.
func init() {
	Register(Spec{
		OracleID:     "0e11792b-7fe5-4208-aa0b-e5d09b2b65fe",
		Name:         "Baneslayer Angel",
		Completeness: CompletenessFull,
		PrintedKeywords: []string{
			"flying", "first strike", "lifelink",
			"protection from Demons", "protection from Dragons",
		},
	})
}
