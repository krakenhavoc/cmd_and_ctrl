package effects

// Iron Spider, Stark Upgrade — Legendary Artifact Creature — Spider
// Hero {3}, 2/3 (EDHREC rank 3195):
//
//	"Vigilance
//	 {T}: Put a +1/+1 counter on each artifact creature and/or
//	 Vehicle you control.
//	 {2}, Remove two +1/+1 counters from among artifacts you
//	 control: Draw a card."
//
// The artifact-creature deck's team pump. Vigilance rides
// PrintedKeywords; the tap ability puts a counter on every artifact
// creature and every Vehicle the controller controls, the Spider
// itself included, snapshotted before the first counter lands. A
// creature source, so the tap waits out summoning sickness (CR
// 302.1), as printed.
//
// One declared simplification, weaker than printed: the draw
// ability is not implemented. "Remove two +1/+1 counters from among
// artifacts you control" is a counter-removal cost spread across
// several permanents, and an ability cost has no counter component
// (AbilityCost carries tap, sacrifice, mana, life, loyalty and crew).
// Shipping the draw without its cost would be stronger than printed
// (#259), so the ability is left off and the counters simply build
// up.
func init() {
	Register(Spec{
		OracleID:        "e123fd7d-ace9-48a4-9510-eedcc837d8e8",
		Name:            "Iron Spider, Stark Upgrade",
		Completeness:    CompletenessCaveats,
		Caveats:         []string{"The draw ability isn't implemented — removing two +1/+1 counters from among your artifacts isn't a cost the engine can pay, so only the tap-to-pump works."},
		PrintedKeywords: []string{"vigilance"},
		Activated: []ActivatedAbility{{
			Label:  "{T}: Put a +1/+1 counter on each artifact creature and/or Vehicle you control",
			Cost:   TapCost(),
			Effect: b30PutCounterOnEachArtifactCreatureOrVehicleYouControl,
		}},
	})
}
