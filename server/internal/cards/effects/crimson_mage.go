package effects

// Crimson Mage — Creature — Human Shaman {1}{R}, 2/1 (Magic 2012, #5573):
//
//	"{R}: Target creature you control gains haste until end of turn.
//	 (It can attack and {T} this turn.)"
//
// Raccoon Rallier's row at a mana cost and instant speed: the target gains
// haste until end of turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d7bf23c4-172a-42fd-b826-3b074a588610",
		Name:         "Crimson Mage",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{R}: Target creature you control gains haste until end of turn.",
			Cost:    ManaCost("{R}"),
			Targets: TargetCreature("target creature you control", YouControl()),
			Purpose: ForTargets(BenefitUntilEOT(0, 0, 0, "haste")),
			Effect:  firstTargetGainsUntilEndOfTurn(0, "Crimson Mage — haste until end of turn", "haste"),
		}},
	})
}
