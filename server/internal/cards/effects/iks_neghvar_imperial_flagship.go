package effects

// I.K.S. Negh'Var, Imperial Flagship — Legendary Artifact — Spacecraft
// {3}, 6/6 (Star Trek Commander, #4974):
//
//	"Klingons and Warriors you control have vigilance.
//	 Station (Tap another creature you control: Put charge counters equal
//	 to its power on this Spacecraft. Station only as a sorcery. It's an
//	 artifact creature at 10+.)
//	 10+ | Flying, vigilance"
//
// The grant is a tribal keyword lord over Klingon and Warrior creatures
// you control, read through the layered view, at every charge-counter
// count. Station is the ordinary keyword (station.go): at ten counters it
// is a 6/6 artifact creature with flying and vigilance.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "373e2b58-ccf0-444b-aeb2-0b6e8d4e80e7",
		Name:         "I.K.S. Negh'Var, Imperial Flagship",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{Station()},
		Static: append(append(
			SpacecraftAt(10, 6, 6),
			ThresholdKeywords(10, "flying", "vigilance")),
			TribalKeywordGrant(TribeFilter{Tribes: []string{"Klingon", "Warrior"}, YoursOnly: true}, "vigilance")),
	})
}
