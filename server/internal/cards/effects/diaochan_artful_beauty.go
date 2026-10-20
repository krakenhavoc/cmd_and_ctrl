package effects

// Diaochan, Artful Beauty — Legendary Creature — Human Advisor {3}{R}, 1/1 (#4062):
//
//	"{T}: Destroy target creature of your choice, then destroy target creature of an opponent's choice.
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
// The first target is yours, the second an opponent's
// (OfAnOpponentsChoice, #2055): the engine asks which opponent, when more
// than one could choose, and that opponent for the creature, before
// anything is paid. Both may be the same creature, and the opponent may
// pick one of yours. The destroys happen in printed order, each skipped if
// its target is gone (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bf20bc37-3205-45af-a3be-05d6835c0d87",
		Name:         "Diaochan, Artful Beauty",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: "{T}: Destroy target creature of your choice, then destroy target creature of an opponent's choice. Activate only during your turn, before attackers are declared.",
			Cost:  TapCost(),
			Targets: Clauses(
				TargetCreature("target creature of your choice"),
				OfAnOpponentsChoice(TargetCreature("target creature of an opponent's choice"))),
			Condition: DuringYourTurnBeforeAttackersAreDeclared(),
			Effect:    diaochanDestroysTwo,
		}},
	})
}
