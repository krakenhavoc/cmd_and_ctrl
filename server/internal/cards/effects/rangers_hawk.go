package effects

// Ranger's Hawk — Creature — Bird {W}, 1/1:
//
//	"Flying
//	 {3}, {T}, Tap another untapped creature you control: Venture into
//	 the dungeon. Activate only as a sorcery."
//
// The {T} waits out summoning sickness (CR 302.6, the engine's check);
// tapping the other creature is a cost that does not (it is not the {T}
// symbol), so a creature that arrived this turn may pay it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "44924f1e-edc4-481b-93f3-78606d4fd691",
		Name:            "Ranger's Hawk",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:        "{3}, {T}, Tap another untapped creature you control: Venture into the dungeon. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{3}"), TapCost(), TapAnotherUntapped("another untapped creature you control", Creature())),
			SorcerySpeed: true,
			Effect:       Venture,
		}},
	})
}
