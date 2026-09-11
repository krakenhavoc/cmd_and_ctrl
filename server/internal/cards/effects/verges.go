package effects

// verges.go — the Duskmourn / Foundations "verge" cycle:
//
//	"{T}: Add {A}."
//	"{T}: Add {B}. Activate only if you control a <X> or a <Y>."
//
// Three of the ten, the three this deck plays.
//
// # Declared sandbox simplification: only the unconditional half
//
// A verge is two separate mana abilities, one of which carries an
// activation restriction. `ManaAbilityCost` has tap / sacrifice /
// sacrifice-another and no condition slot, and `ActivateManaAbility`
// has no hook to evaluate one, so the second ability cannot be
// declared without an engine change.
//
// So these ship with ONLY their unconditional ability. That makes
// each verge a strictly worse card than the printed one — a
// mono-coloured land — which is the right direction for a
// simplification to point: a player can never get mana the real card
// would not have given them. The cost is that a verge does not fix
// colours, which is its entire job.
//
// The fix is small and shared: a condition on ManaAbilityCost,
// evaluated at activation time, plus the same filter hook
// `commanderIdentityFor` already uses to narrow Arcane Signet's pipe
// set. It would also unblock Mox Amber ("any colour among legendary
// creatures and planeswalkers you control") and half of Chrome Mox.
// Until then the restricted half is absent rather than free.
func init() {
	// `free` is the colour of the unconditional ability — the only
	// half that ships.
	for _, t := range []struct {
		oracleID string
		name     string
		free     string
	}{
		{"2b8144a0-08d2-4c28-9fd7-5d90f90105e4", "Bleachbone Verge", "B"},
		{"f1e9abfb-c3c8-483e-b446-5c2afc9f6394", "Floodfarm Verge", "W"},
		{"d71bda4c-3dee-4398-8fd0-f77d8743b887", "Gloomlake Verge", "U"},
	} {
		Register(Spec{
			OracleID:     t.oracleID,
			Name:         t.name,
			Completeness: CompletenessCaveats,
			Caveats:      []string{"Only the unconditional half is available — the second, conditional color the verge can make is missing, so it taps for one color."},
			ManaAbilities: []ManaAbility{{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{" + t.free + "}",
				Label:    "Add {" + t.free + "}",
			}},
		})
	}
}
