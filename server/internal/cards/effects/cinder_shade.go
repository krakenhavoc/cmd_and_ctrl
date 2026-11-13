package effects

// Cinder Shade — Creature — Shade {1}{B}{R}, 1/1 (Invasion, #5616):
//
//	"{B}: This creature gets +1/+1 until end of turn.
//	 {R}, Sacrifice this creature: It deals damage equal to its power to
//	 target creature."
//
// The pump is the shared thisGetsRow; the sacrifice is Heartfire
// Immolator's sacrificedSourceDealsItsPowerToTheTarget. The sacrifice is
// a cost, so "its power" is its power as it last existed on the
// battlefield (CR 113.7a), every {B} pump included. Neither cost has
// {T}, so both work the turn the Shade arrives.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9e120284-0fcc-432c-9471-9106eab71025",
		Name:         "Cinder Shade",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			thisGetsRow("{B}", 1, 1, "Cinder Shade"),
			{
				Label:   "{R}, Sacrifice this creature: It deals damage equal to its power to target creature.",
				Cost:    Plus(ManaCost("{R}"), SacrificeThis()),
				Targets: TargetCreature("target creature"),
				Effect:  sacrificedSourceDealsItsPowerToTheTarget,
			},
		},
	})
}
