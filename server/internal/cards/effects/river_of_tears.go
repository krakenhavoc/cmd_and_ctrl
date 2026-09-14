package effects

// River of Tears — Land (EDHREC rank 2698):
//
//	"{T}: Add {U}. If you played a land this turn, add {B} instead."
//
// The Future Sight dual that is blue on an opponent's turn and black
// on yours once the land drop is made. One tap ability whose colour
// is derived at activation (ProducedFunc, the Cabal Coffers shape)
// from the engine's per-turn land-play tally: a land played this
// turn — this one included, if it was the drop — makes it {B}, and
// anything else makes it {U}. The tally is per player and resets
// every turn, so on another player's turn the River is blue.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8a83d284-75a0-4901-b7d9-c4b7586ee327",
		Name:         "River of Tears",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:         ManaAbilityCost{Tap: true},
			ProducedFunc: b25BlackIfLandPlayedElseBlue,
			Label:        "{T}: Add {U}, or {B} if you played a land this turn",
		}},
	})
}
