package effects

// Oakshade Stalker // Moonlit Ambusher — {2}{G} Creature — Human Ranger
// Werewolf 3/3 // Creature — Werewolf 6/3 (#2586, ADR 0132):
//
//	Front: "You may cast this spell as though it had flash if you pay {2}
//	        more to cast it.
//	        Daybound"
//	Back:  "Nightbound"
//
// Not implemented: the flash-for-{2}-more option. An alternative cost
// carries a price but no timing, and a spell's own timing is a keyword
// the card either has or lacks, so there is no way yet to say "has flash
// only if this surcharge was paid". The card is cast at sorcery speed
// for {2}{G} and nothing else about it is missing — the weaker
// direction, which is why it ships with a caveat rather than a flash it
// would grant for free.
func init() {
	const oracle = "1b639537-a8fe-4615-b937-afa478fdec0f"
	Register(Spec{
		OracleID:        oracle,
		Name:            "Oakshade Stalker",
		Completeness:    CompletenessCaveats,
		Caveats:         []string{"It can't be cast as though it had flash for {2} more — it can only be cast when you could cast a sorcery."},
		PrintedKeywords: []string{"daybound"},
	})
	Register(Spec{
		OracleID:        oracle + "#1",
		Name:            "Moonlit Ambusher",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"nightbound"},
	})
}
