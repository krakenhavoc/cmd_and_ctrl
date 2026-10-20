package effects

// Hua Tuo, Honored Physician — Legendary Creature — Human {1}{G}{G}, 1/2 (#4062):
//
//	"{T}: Put target creature card from your graveyard on top of your library.
//	 Activate only during your turn, before attackers are declared."
//
// The window is DuringYourTurnBeforeAttackersAreDeclared, an activation
// condition (CR 602.1b): your own turn, and only until the first declare
// attackers step of the turn begins - the card's ruling reads it from the
// first combat, so a turn with several combat phases does not open it
// again. Not a sorcery-speed gate: it opens your beginning, combat and
// main-phase steps up to that point, with or without an item on the stack,
// and the view and the legal enumerator read the same closure, so the
// ability greys out when the window shuts.
//
// The card is tucked through the shared exit primitive, so a commander
// gets its CR 903.9 offer on the way.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "00284955-0a1b-4939-9567-a7d34cbcee51",
		Name:         "Hua Tuo, Honored Physician",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     "{T}: Put target creature card from your graveyard on top of your library. Activate only during your turn, before attackers are declared.",
			Cost:      TapCost(),
			Targets:   TargetCardInGraveyard("target creature card from your graveyard", Creature(), YouOwn()),
			Condition: DuringYourTurnBeforeAttackersAreDeclared(),
			Effect:    PutChosenTargetOnTopOfLibrary,
		}},
	})
}
