package effects

// Time Stop — {4}{U}{U} Instant:
//
//	"End the turn. (Exile all spells and abilities, including this
//	 spell. The player whose turn it is discards down to their maximum
//	 hand size. Damage heals and "this turn" and "until end of turn"
//	 effects end.)"
//
// EndTheTurn (end_the_turn.go, CR 724.1, #2165). Cast on any turn: on
// an opponent's it exiles whatever they had on the stack — a
// "can't be countered" spell included, because exiling is not
// countering — takes their attackers out of combat before damage, and
// sends their turn to its cleanup step. Time Stop exiles itself with
// the rest of the stack (CR 724.1b), so it never reaches a graveyard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "55bc6a7e-2356-48f0-a93a-8bf19044ee4b",
		Name:         "Time Stop",
		Completeness: CompletenessFull,
		OnResolve:    endTheTurnOnResolve,
	})
}
