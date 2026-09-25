package effects

// Windgrace's Judgment — Instant {3}{B}{G} (EDHREC rank 3173):
//
//	"For any number of opponents, destroy target nonland permanent
//	 that player controls."
//
// One removal spell per opponent at instant speed — the multiplayer
// Putrefy. The target clause is "any number of target nonland
// permanents your opponents control", and since #1559 the "one per
// opponent" is the clause's own set rule, EachDifferentController:
// the announce gate refuses two permanents of one opponent's
// (CR 601.2c), the picker greys the second, and the bot is never
// offered the pair. Before #1559 the picker accepted the pair and the
// second one was silently skipped as the spell resolved.
//
// One declared simplification is left, and it is narrow. The printed
// clause binds each target to the opponent it was chosen for, so a
// permanent that changes control in response is an illegal target
// (CR 608.2b). The set rule has no memory of that binding: it judges
// the targets by who controls them as the spell resolves. A permanent
// that moved to an opponent you did not target is still destroyed,
// and two targets that end up under one controller are both spared.
// Binding a pick to a player would need a per-target record on the
// stack item that nothing else needs yet; Molten Primordial, whose
// clause is the same sentence on a trigger, gets the exact reading
// from one clause per opponent (molten_primordial.go), which a spell's
// static clause list cannot build.
func init() {
	Register(Spec{
		OracleID:     "de1ca6ed-b275-4f62-ba05-f31b3659b352",
		Name:         "Windgrace's Judgment",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"A permanent that changes control before the spell resolves is judged by its new controller, not by the opponent you chose it for."},
		Targets: TargetPermanent("any number of target nonland permanents your opponents control, one per opponent", Nonland(), OpponentControls()).
			WithCount(1, 0).EachDifferent(EachDifferentController()),
		OnResolve: destroyEachLegalTarget,
	})
}
