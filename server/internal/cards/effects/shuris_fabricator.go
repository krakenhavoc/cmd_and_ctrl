package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shuri's Fabricator — {4} Artifact (Marvel Super Heroes Commander, #4455):
//
//	"When this artifact enters, create two tapped Vibranium tokens.
//	 {6}, {T}: Return target artifact card from your graveyard to the
//	 battlefield with a finality counter on it. Activate only as a
//	 sorcery. (If a permanent with a finality counter on it would be put
//	 into a graveyard, exile it instead.)"
//
// The enters trigger is Dora Milaje Elite's token with N: 2. The return is
// Grim Reaper's Scythe's, over an artifact card: the card comes back with
// the finality counter as it enters, and CR 122.1h's replacement
// (game/finality_counters.go) is the engine's. The target is chosen on
// activation and the card must still be in the graveyard on resolution.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "127f6985-c1f4-487f-b5ee-c84524cd6c55",
		Name:         "Shuri's Fabricator",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Shuri's Fabricator — create two tapped Vibranium tokens",
				Do(CreateTokenAdvanced{Spec: Token(VibraniumToken()).EntersTapped(), N: 2})),
		},
		Activated: []ActivatedAbility{{
			Label:        "{6}, {T}: Return target artifact card from your graveyard to the battlefield with a finality counter on it. Activate only as a sorcery.",
			Purpose:      ForTargets(Reanimates(0)),
			Cost:         Plus(ManaCost("{6}"), TapCost()),
			SorcerySpeed: true,
			Targets:      TargetCardInGraveyard("target artifact card in your graveyard", Artifact(), YouOwn()),
			Effect:       grimReapersScytheReturn,
		}},
	})
}
