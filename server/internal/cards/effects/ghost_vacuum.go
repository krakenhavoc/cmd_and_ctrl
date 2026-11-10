package effects

// Ghost Vacuum — Artifact {1} (Duskmourn: House of Horror, #4466):
//
//	"{T}: Exile target card from a graveyard.
//	 {6}, {T}, Sacrifice this artifact: Put each creature card exiled with
//	 this artifact onto the battlefield under your control with a flying
//	 counter on it. Each of them is a 1/1 Spirit in addition to its other
//	 types. Activate only as a sorcery."
//
// The two abilities are linked (CR 607.2a): the first exiles with the
// link stamped on each card (ExileTarget.Linked), the second reads the cards
// linked to the object the ability came from, which survives the
// artifact's own sacrifice. The creatures land as one entry with a flying
// counter, and the "1/1 Spirit in addition to its other types" is an entry
// overwrite on each (EntersAs, Excava, the Risen Past's shape: a layer 4
// subtype add and layer 7b base power and toughness, ending when the
// object leaves).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4ef583ab-e0a0-4e4b-bd65-af766038b1fd",
		Name:         "Ghost Vacuum",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{T}: Exile target card from a graveyard.",
				Cost:    TapCost(),
				Targets: TargetCardInGraveyard("target card from a graveyard"),
				Effect:  stdB26GhostVacuumExile,
			},
			{
				Label: "{6}, {T}, Sacrifice this artifact: Put each creature card exiled with this artifact onto the battlefield under your control " +
					"with a flying counter on it. Each of them is a 1/1 Spirit in addition to its other types. Activate only as a sorcery.",
				Cost:         Plus(ManaCost("{6}"), TapCost(), SacrificeThis()),
				SorcerySpeed: true,
				Effect:       stdB26GhostVacuumReanimate,
			},
		},
	})
}
