package effects

// Deadeye Duelist — Creature — Human Assassin {1}{R}, 1/3 (Outlaws of
// Thunder Junction, #4468):
//
//	"Reach
//	 {1}, {T}: This creature deals 1 damage to target opponent."
//
// Fervent Paincaster's pinger with a {1} on the cost and the target narrowed
// to an opponent. The damage is dealt by the creature, and the opponent is
// checked again as the ability resolves (CR 608.2b). Targeting an opponent
// is a crime (CR 700.13), which the engine reads from the activation.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "29ef90ac-2396-4488-970d-d440297ba5ee",
		Name:            "Deadeye Duelist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}: This creature deals 1 damage to target opponent.",
			Cost:    Plus(ManaCost("{1}"), TapCost()),
			Targets: TargetPlayer("target opponent", Opponent()),
			Purpose: ForTargets(DamageToTarget(0, 1)),
			Effect:  sourceDealsDamageToEachLegalTarget(1),
		}},
	})
}
