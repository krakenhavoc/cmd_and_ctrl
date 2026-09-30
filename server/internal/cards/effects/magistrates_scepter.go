package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Magistrate's Scepter — Artifact {3}:
//
//	"{4}, {T}: Put a charge counter on this artifact.
//	 {T}, Remove three charge counters from this artifact: Take an
//	 extra turn after this one."
//
// Lux Cannon's shape with a different payoff. Both abilities tap it, so
// charging and firing take different turns unless something untaps it.
// Removing the three counters is a COST paid at announce (#625), so a
// response cannot stop the turn by removing counters. The turn is
// TakeExtraTurn (CR 500.7, ADR 0059 Decision 5) and goes to the
// ability's controller.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "487485a6-cf79-48c0-bea9-0ec6b3ee253c",
		Name:         "Magistrate's Scepter",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:  "{4}, {T}: Put a charge counter on this artifact.",
				Cost:   Plus(ManaCost("{4}"), TapCost()),
				Effect: b33PutChargeCounterOnSelf,
			},
			{
				Label:  "{T}, Remove three charge counters from this artifact: Take an extra turn after this one.",
				Cost:   Plus(TapCost(), RemoveCountersFromThis(game.CounterCharge, 3)),
				Effect: youTakeAnExtraTurnEffect,
			},
		},
	})
}
