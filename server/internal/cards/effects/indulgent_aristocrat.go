package effects

// Indulgent Aristocrat — Creature — Vampire Noble {B}, 1/1 (EDHREC
// rank 2555):
//
//	"Lifelink
//	 {2}, Sacrifice a creature: Put a +1/+1 counter on each Vampire
//	 you control."
//
// The one-drop Vampire sac outlet. Lifelink rides PrintedKeywords;
// the activation is a CR 602 ability with a mana-and-sacrifice cost
// (the Aristocrat itself is a legal sacrifice — it is a creature —
// and then grows nothing), and its body is Cordial Vampire's
// b17PutCounterOnEachVampireYouControl: every Vampire the controller
// controls, the Aristocrat included, post-layer subtype so a
// changeling counts, snapshotted before the first counter lands.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c7054a2d-2e7b-4487-8fc3-f6a47a716fd3",
		Name:            "Indulgent Aristocrat",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Activated: []ActivatedAbility{{
			Label:  "{2}, Sacrifice a creature: Put a +1/+1 counter on each Vampire you control.",
			Cost:   Plus(ManaCost("{2}"), SacrificeACreature()),
			Effect: b17PutCounterOnEachVampireYouControl,
		}},
	})
}
