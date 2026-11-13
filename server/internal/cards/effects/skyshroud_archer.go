package effects

// Skyshroud Archer — Creature — Elf Archer {G}, 1/1 (Stronghold, #5627):
//
//	"{T}: Target creature with flying gets -1/-1 until end of turn."
//
// Centaur Archer's "target creature with flying" clause and Grandmother
// Sengir's shared boostTargetUntilEOT body at -1/-1, ending in the cleanup
// step (CR 514.2). The tap symbol means it can't be activated the turn
// the Archer arrives unless it has haste (CR 302.6). A flier that loses
// flying in response is no longer a legal target and is unchanged
// (CR 608.2b); one left at 0 toughness dies (CR 704.5f).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "42641a05-a6b6-4856-a72c-54181b97c190",
		Name:         "Skyshroud Archer",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: Target creature with flying gets -1/-1 until end of turn.",
			Purpose: ForTargets(HarmUntilEOT(0, -1, -1)),
			Cost:    TapCost(),
			Targets: TargetCreature("target creature with flying", HasKeyword("flying")),
			Effect:  boostTargetUntilEOT("Skyshroud Archer — -1/-1", -1, -1),
		}},
	})
}
