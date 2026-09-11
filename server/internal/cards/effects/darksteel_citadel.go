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
// DECLARED SIMPLIFICATION — INDESTRUCTIBLE IS INERT TODAY. The
// keyword is declared in PrintedKeywords so it lands in
// Card.Effective().Abilities and reads correctly to anything
// inspecting the card, but the engine does not yet ENFORCE it:
// game/keywords.go honours twelve combat keywords (flying, reach,
// first strike, double strike, deathtouch, lifelink, trample,
// vigilance, menace, defender, haste, flash) and indestructible is
// not among them, so DestroyPermanentForEffect routes the Citadel to
// the graveyard like any other permanent. That is #176.
//
// This ships the card WEAKER than printed, never stronger, which is
// the acceptable direction. It is declared here rather than left
// implicit because #350 is the ticket about simplification notes
// outliving their blockers: when #176 lands, this note is the thing
// to delete, and nothing else on the card needs to change — the
// keyword is already declared, so enforcement picks it up for free.
//
// Compare Overrun's file, which declined to ship Heroic Intervention
// because EVERY word of that card would have been a declared no-op.
// Two of Darksteel Citadel's three lines — the artifact land type and
// the mana ability — are fully live.
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
