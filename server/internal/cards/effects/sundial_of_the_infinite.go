package effects

// Sundial of the Infinite — {2} Artifact:
//
//	"{1}, {T}: End the turn. Activate only during your turn. (Exile all
//	 spells and abilities from the stack. Discard down to your maximum
//	 hand size. Damage wears off, and "this turn" and "until end of
//	 turn" effects end.)"
//
// The ability is EndTheTurn (end_the_turn.go, CR 724.1, #2165): the
// stack is exiled — the classic play is in response to your own
// end-step or upkeep trigger, and a "sacrifice it at the beginning of
// the next end step" never fires, because the end step is skipped and
// the trigger waits for the next one — every creature leaves combat,
// and the turn goes straight to its cleanup step.
//
// "Activate only during your turn" is DuringYourTurn(): any step of
// your turn, with or without something on the stack (Sundial's ability
// is not sorcery-speed). The ability itself is on the stack when it
// resolves and is not exiled by its own resolution — it has already
// left the stack — so ending the turn with Sundial's ability is not a
// special case.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "79d46e09-2548-440f-9c02-3c3dc39a0cd1",
		Name:         "Sundial of the Infinite",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     "{1}, {T}: End the turn. Activate only during your turn.",
			Cost:      Plus(ManaCost("{1}"), TapCost()),
			Condition: DuringYourTurn(),
			Effect:    endTheTurnEffect,
		}},
	})
}
