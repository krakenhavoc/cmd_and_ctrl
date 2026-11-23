package effectstest

// pinned_retarget_test.go — #1743's card proofs: Spellskite, Mizzium
// Meddler and Hydroelectric Specimen, each "change a target of target
// spell or ability TO THIS CREATURE". The engine half (which slots are
// eligible, the which-one prompt) is game/retarget_pinned_test.go; this
// is the cards doing it on a real board.

const (
	SpellskiteOracle     = "e0aa6ce0-ca31-433b-ac6c-32b8675cdb71"
	MizziumMeddlerOracle = "48a909b6-e6ee-4148-8b50-b35f11bc065f"
)
