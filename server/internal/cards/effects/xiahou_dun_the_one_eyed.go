package effects

// Xiahou Dun, the One-Eyed — Legendary Creature — Human Soldier {2}{B}{B}, 3/2 (#4062):
//
//	"Sacrifice Xiahou Dun: Return target black card from your graveyard to your hand.
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
// Horsemanship is declared as a printed keyword. Xiahou Dun is a black
// card himself, but he is sacrificed as a cost after the target is chosen,
// so he cannot return himself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9602f262-62cb-4c1e-91b6-d94780e571cf",
		Name:            "Xiahou Dun, the One-Eyed",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"horsemanship"},
		Activated: []ActivatedAbility{{
			Label:     "Sacrifice Xiahou Dun: Return target black card from your graveyard to your hand. Activate only during your turn, before attackers are declared.",
			Cost:      SacrificeThis(),
			Targets:   TargetCardInGraveyard("target black card from your graveyard", OfColor("B"), YouOwn()),
			Condition: DuringYourTurnBeforeAttackersAreDeclared(),
			Effect:    b27ReturnChosenGraveyardCardToHand,
		}},
	})
}
