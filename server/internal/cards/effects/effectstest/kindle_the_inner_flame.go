package effectstest

// kindle_the_inner_flame_test.go — #4689: "Flashback—{1}{R}, Behold three
// Elementals" (CR 702.34a, CR 701.4a). The flashback cost names three
// distinct objects on alt_cost_ids, each an Elemental the caster controls
// or an Elemental card in their hand; the hand cards are revealed, nothing
// moves, and the card is exiled as it leaves the stack.

const (
	KindleOracle    = "c7ca6371-73bb-4717-a148-0804b8f4c1e6"
	KindleName      = "Kindle the Inner Flame"
	KindleTypeLine  = "Kindred Sorcery — Elemental"
	ElementalType   = "Creature — Elemental"
	KindleFlashback = "flashback"
)
