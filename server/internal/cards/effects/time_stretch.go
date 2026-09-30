package effects

// Time Stretch — Sorcery {8}{U}{U}:
//
//	"Target player takes two extra turns after this one."
//
// Two turns from one effect, queued together (CR 500.7, ADR 0059
// Decision 5): both go ahead of any extra turn queued earlier, and an
// extra turn created DURING the first of them is taken before the
// second — the order the card's rulings give.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "72e56963-a9dd-44dc-a4d3-992b4d89dd28",
		Name:         "Time Stretch",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		OnResolve:    targetPlayerTakesExtraTurns(2),
	})
}
