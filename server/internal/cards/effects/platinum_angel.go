package effects

// Platinum Angel — Artifact Creature — Angel {7}, 4/4:
//
//	"Flying
//	 You can't lose the game and your opponents can't win the game."
//
// The reference card for ADR 0057 Decision 4 (#749). The static is
// declared, not written: Spec.GameEndGates, read off the battlefield
// through game.CatalogGameEndGates at every state-based loss check, at
// every effect loss and at every effect win. Its 2009 ruling is the
// whole contract — "No game effect can cause you to lose the game or
// cause any opponent to win the game while you control Platinum
// Angel", whatever the route: life, an empty library, poison, commander
// damage, an effect. A concession still loses (CR 104.3a), and the
// last player standing still wins (CR 104.2a overrides "can't win").
//
// Two Angels compose, one leaving can't revoke the other's gate, an
// Angel under Humility or Dress Down gates nothing, and an Angel that
// changes controller protects its new controller — all because the
// gate is derived rather than stored. When it leaves, the next check
// reads life, poison and commander damage afresh, so a player it was
// keeping alive at −5 loses then, with no window.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:        "b148578c-c0bf-4785-b97c-4b6f83028008",
		Name:            "Platinum Angel",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		GameEndGates:    YouCantLoseOpponentsCantWin(),
	})
}
