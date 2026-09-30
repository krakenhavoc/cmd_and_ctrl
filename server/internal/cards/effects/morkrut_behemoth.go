package effects

// Morkrut Behemoth — Creature — Zombie Giant {4}{B}, 7/6:
//
//	"As an additional cost to cast this spell, sacrifice a creature or
//	 pay {1}{B}. Menace"
//
// The either/or cost (ADR 0100 §2) on a creature spell; the {1}{B}
// branch joins the total at CR 601.2f through the one pricer.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "76c2571d-7d45-4c7a-99d1-302e2b26f8f6",
		Name:         "Morkrut Behemoth",
		Completeness: CompletenessFull,
		AdditionalCost: EitherCost(
			SacrificeCost("a creature", Creature()).Keyed("sacrifice"),
			ManaAdditionalCost("{1}{B}").Keyed("mana"),
		),
		PrintedKeywords: []string{"menace"},
	})
}
