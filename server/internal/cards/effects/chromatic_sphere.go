package effects

// Chromatic Sphere — Artifact {1}:
//
//	"{1}, {T}, Sacrifice this artifact: Add one mana of any color.
//	 Draw a card. (Activate only as an instant.)"
//
// Chromatic Star's sibling with the draw on the ability instead of on
// leaving the battlefield. It adds mana, so it is a mana ability (CR
// 605.1a) and the "Draw a card" is a rider on it (samiDrawRider): the
// mana lands, then the card is drawn, in one activation with no stack.
// All three cost components are validated before any is paid. "Activate
// only as an instant" is every activated ability's default timing.
//
// A hand-written rider is opaque to the auto-tapper, so the Sphere is
// never tapped for you to pay a cast; activate it yourself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2e03e44a-9fff-4490-859f-b42e89e8563a",
		Name:         "Chromatic Sphere",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			samiAnyColorManaThenDraw(ManaAbilityCost{Tap: true, Sacrifice: true, Mana: "{1}"},
				"{1}, {T}, Sacrifice: Add one mana of any color. Draw a card"),
		},
	})
}
