package effects

// Hydroelectric Specimen // Hydroelectric Laboratory — the FRONT
// face, Creature — Weird {2}{U}, 1/4:
//
//	"Flash
//	 When this creature enters, you may change the target of target
//	 instant or sorcery spell with a single target to this creature."
//
// The land back (pay 3 life or enter tapped; {T}: Add {U}) is the
// mdfc_lands.go row under "<oracle>#1"; this is face 0, which keeps
// the bare oracle ID (game.CatalogKey). Flash rides PrintedKeywords.
//
// The ETB is left unbuilt. CR 115.7's engine primitive
// (effects.ChangeTargets, retarget.go) is "change the target of …",
// which always opens the RETARGETING PLAYER a picker over every
// legal new target (Bolt Bend, Deflecting Swat) — CR 115.7b/c. This
// card's clause has no picker at all: the new target is fixed, "to
// this creature", with no choice offered. Building it on top of
// ChangeTargets would let the controller point the spell anywhere
// legal, which is STRONGER than printed (#259's wrong direction), so
// it is left out rather than shipped wrong. What's missing is a
// retarget primitive whose new target is a fixed object rather than a
// chooser's pick — filed as a new seam alongside this slice's PR
// (retarget-to-a-fixed-object).
func init() {
	Register(Spec{
		OracleID:        "573151f0-00d4-4a8a-8a09-745c5f376532",
		Name:            "Hydroelectric Specimen",
		Completeness:    CompletenessCaveats,
		Caveats:         []string{"The enter-the-battlefield ability isn't implemented — this creature can't redirect a targeted instant or sorcery spell to itself."},
		PrintedKeywords: []string{"flash"},
	})
}
