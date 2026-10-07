package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dynavolt Tower — Artifact {3}:
//
//	"Whenever you cast an instant or sorcery spell, you get {E}{E} (two
//	 energy counters).
//	 {T}, Pay {E}{E}{E}{E}{E}: This artifact deals 3 damage to any
//	 target."
//
// ADR 0129 §2 (#1995): "Pay {E}{E}{E}{E}{E}" is the energy cost
// component. The cast trigger resolves before the spell that caused it.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "59f56dd9-cc82-4f0d-a522-0d8ed274fa40",
		Name:         "Dynavolt Tower",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(WheneverYouCast(Or(Instant(), Sorcery()), "Dynavolt Tower — you get {E}{E}", Do(GetEnergy{N: 2})),
				game.Purpose{Energy: 2}),
		},
		Activated: []ActivatedAbility{{
			Label:   "{T}, Pay {E}{E}{E}{E}{E}: This artifact deals 3 damage to any target.",
			Cost:    Plus(TapCost(), PayEnergy(5)),
			Targets: TargetAny(),
			Effect:  sourceDealsDamageToEachLegalTarget(3),
		}},
	})
}
