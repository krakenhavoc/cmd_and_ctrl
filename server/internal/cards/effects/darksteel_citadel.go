package effects

// Darksteel Citadel — Artifact Land:
//
//	"Indestructible"
//	"{T}: Add {C}."
//
// An artifact land that survives a board wipe. Like Seat of the
// Synod it is played for its type line — it is an ARTIFACT, so it
// feeds metalcraft, affinity and every artifact count — and unlike
// Seat of the Synod it also survives the sweepers those decks fear.
//
// Registered for the usual nonbasic-land reason: the synthetic mana
// ability only fires for lands with the BASIC supertype, so an
// unregistered Citadel taps for nothing.
//
// INDESTRUCTIBLE IS LIVE AS OF S25 (#77). This file used to carry a
// declared simplification saying the keyword was inert — declared in
// PrintedKeywords so it read correctly to anything inspecting the
// card, but enforced nowhere, so DestroyPermanentForEffect routed
// the Citadel to the graveyard like any other permanent. S25 taught
// the destruction path CR 702.12 (server/internal/game/
// indestructible.go), and the prediction that note made came true
// exactly: nothing on this card changed. The keyword was already
// declared, so enforcement picked it up for free, and deleting the
// note was the entire diff.
//
// All three of the Citadel's lines are now live: the artifact land
// type, the mana ability, and indestructible. It really does survive
// the sweepers its decks fear.
func init() {
	Register(Spec{
		OracleID:        "8dc067bf-f78f-4ac4-b6e7-b305c42cf0bc",
		Name:            "Darksteel Citadel",
		PrintedKeywords: []string{"indestructible"},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
	})
}
