package effects

// Bosh, Iron Golem — Legendary Artifact Creature — Golem {8}, 6/7
// (Mirrodin, #5606):
//
//	"Trample
//	 {3}{R}, Sacrifice an artifact: Bosh deals damage equal to the
//	 sacrificed artifact's mana value to any target."
//
// Pia and Kiran Nalaar's sacrifice-an-artifact cost with Broadside
// Bombardiers' damage: the amount is the sacrificed artifact's mana
// value as it last existed on the battlefield (sacrificedManaValue,
// CR 608.2h), so a token that is not a copy deals none. Bosh is an
// artifact and may sacrifice itself; the damage then comes from Bosh's
// last known information.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2fb6f65f-a0fb-4d45-be1b-d6405e61d47a",
		Name:            "Bosh, Iron Golem",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Activated: []ActivatedAbility{{
			Label:   "{3}{R}, Sacrifice an artifact: Bosh deals damage equal to the sacrificed artifact's mana value to any target.",
			Cost:    Plus(ManaCost("{3}{R}"), SacrificeN(1, "an artifact", Artifact())),
			Targets: TargetAny(),
			Effect:  sourceDealsComputedDamageToEachLegalTarget(sacrificedManaValue),
		}},
	})
}
