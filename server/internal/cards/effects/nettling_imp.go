package effects

// Nettling Imp — Creature — Imp {2}{B}, 1/1 (#4062):
//
//	"{T}: Choose target non-Wall creature the active player has
//	 controlled continuously since the beginning of the turn. That
//	 creature attacks this turn if able. Destroy it at the beginning of
//	 the next end step if it didn't attack this turn. Activate only
//	 during an opponent's turn, before attackers are declared."
//
// The window is DuringAnOpponentsTurnBeforeAttackersAreDeclared: any
// other player's turn, until the first declare attackers step begins. The
// target is a non-Wall creature the ACTIVE player has controlled since
// the turn began (the summoning-sickness reading, CR 302.6), so a
// creature that arrived this turn is not a legal pick. It gets the
// "attacks this turn if able" requirement until end of turn (CR 508.1d)
// and a delayed trigger that destroys it at the beginning of the next end
// step unless it was declared as an attacker, whatever the reason it
// didn't (CR 603.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c58dfcbf-49e6-4ef0-bd31-ebd81b0cfa41",
		Name:         "Nettling Imp",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     "{T}: Choose target non-Wall creature the active player has controlled continuously since the beginning of the turn. That creature attacks this turn if able. Destroy it at the beginning of the next end step if it didn't attack this turn. Activate only during an opponent's turn, before attackers are declared.",
			Cost:      TapCost(),
			Targets:   forcedAttackTarget(),
			Condition: DuringAnOpponentsTurnBeforeAttackersAreDeclared(),
			Effect:    forceTheChosenCreature("Nettling Imp"),
		}},
	})
}
