package effects

// Etherium Abomination — Artifact Creature — Horror {3}{U}{B}, 4/3 (Alara
// Reborn, #5586):
//
//	"Unearth {1}{U}{B} ({1}{U}{B}: Return this card from your graveyard to
//	 the battlefield. It gains haste. Exile it at the beginning of the next
//	 end step or if it would leave the battlefield. Unearth only as a
//	 sorcery.)"
//
// The shared Unearth row (CR 702.84a, unearth.go), as on Hell's Thunder.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e1135564-ca93-4c5b-8686-cafdbaf0205e",
		Name:         "Etherium Abomination",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{Unearth("{1}{U}{B}")},
	})
}
