package effects

// Alms of the Vein — Sorcery {2}{B} (Shadows over Innistrad, #5543):
//
//	"Target opponent loses 3 life and you gain 3 life.
//	 Madness {B} (If you discard this card, discard it into exile. When
//	 you do, cast it for its madness cost or put it into your graveyard.)"
//
// Brush with Death's drain at three (targetPlayerLosesLifeYouGain), aimed
// at an opponent. Madness is grown by the engine from the one string (CR
// 702.35a, #657).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2013a774-e6ab-4f2d-a84f-fd277a637d7e",
		Name:         "Alms of the Vein",
		Completeness: CompletenessFull,
		Madness:      "{B}",
		Targets:      TargetPlayer("target opponent", Opponent()),
		OnResolve:    targetPlayerLosesLifeYouGain(3),
	})
}
