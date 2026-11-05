package effects

// Scotty, Chief Engineer — {3}{U} Legendary Creature — Human Artificer
// 3/4 (Star Trek Commander, #4974):
//
//	"Tap an untapped artifact you control: Target creature can't be
//	 blocked this turn."
//
// The cost is TapAnUntapped over artifacts (Urza, Lord High Artificer's
// cost on a non-mana ability): not the {T} symbol, so an artifact that
// arrived this turn may pay it (CR 302.6), and it does not target. Scotty
// is not an artifact, so he never pays it himself. The effect is Zhuge
// Jin's targetCantBeBlockedThisTurn; a target gone by resolution gets
// nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3bd553db-2c7d-42cd-8ded-ab2c955b8f4e",
		Name:         "Scotty, Chief Engineer",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Tap an untapped artifact you control: Target creature can't be blocked this turn.",
			Purpose: ForTargets(CantBeBlockedUntilEOT(0, 0, 0)),
			Cost:    TapAnUntapped("an untapped artifact you control", Artifact()),
			Targets: TargetCreature("target creature"),
			Effect:  targetCantBeBlockedThisTurn("Scotty, Chief Engineer"),
		}},
	})
}
