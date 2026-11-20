package effects

// Emberwilde Augur — Creature — Goblin Shaman {1}{R}, 2/1 (Future Sight, #5594):
//
//	"Sacrifice this creature: It deals 3 damage to target player or
//	 planeswalker. Activate only during your upkeep."
//
// An upkeep-only sacrifice (DuringYourUpkeep is the step AND your turn).
// The sacrificed creature is the source of the damage, read from its
// last-known information once it has left (CR 608.2h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "39bf60b2-eeee-48f3-b7b0-6cc2e038afb1",
		Name:         "Emberwilde Augur",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     "Sacrifice this creature: It deals 3 damage to target player or planeswalker. Activate only during your upkeep.",
			Purpose:   ForTargets(DamageToTarget(0, 3)),
			Cost:      SacrificeThis(),
			Condition: DuringYourUpkeep(),
			Targets:   targetPlayerOrPlaneswalker(),
			Effect:    sourceDealsDamageToEachLegalTarget(3),
		}},
	})
}
