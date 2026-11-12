package effects

// Femeref Archers — Creature — Human Archer {2}{G}, 2/2 (Mirage, #5633):
//
//	"{T}: This creature deals 4 damage to target attacking creature with
//	 flying."
//
// The target clause is an attacking creature (AttackingCreature) that has
// flying as the ability is activated and again as it resolves (CR 608.2b),
// so a creature that lost flying or was removed from combat in response
// is no longer a legal target and takes nothing. The damage is the shared
// DealDamageToTheTarget body, dealt by the Archers. The tap means
// summoning sickness applies (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b5867796-5898-4a75-a75a-c802407ab338",
		Name:         "Femeref Archers",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: This creature deals 4 damage to target attacking creature with flying.",
			Purpose: ForTargets(DamageToTarget(0, 4)),
			Cost:    TapCost(),
			Targets: TargetCreature("target attacking creature with flying", AttackingCreature(), HasKeyword("flying")),
			Effect:  DealDamageToTheTarget(4),
		}},
	})
}
