package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Simulacrum Synthesizer — Artifact {2}{U} (EDHREC rank 1718):
//
//	"When this artifact enters, scry 2.
//	 Whenever another artifact you control with mana value 3 or
//	 greater enters, create a 0/0 colorless Construct artifact
//	 creature token with "This token gets +1/+1 for each artifact
//	 you control.""
//
// The artifact deck's Construct factory: every real rock or
// equipment after it is a Karnstruct. The entry scry is Temple's;
// the Construct trigger reads the entering artifact's mana value off
// the permanent (b15AnotherBigArtifactYouControlEntered — a token
// with no cost is zero, so Treasures make nothing).
//
// The Construct is Urza's (UrzaConstructToken): the "+1/+1 for each
// artifact you control" is the token's OWN static (ADR 0083), so it keeps
// sizing itself after the Synthesizer leaves and is never smaller than
// 1/1, as printed.
func init() {
	Register(Spec{
		OracleID:     "eb7a1f21-a66d-415b-8520-710b44890bb6",
		Name:         "Simulacrum Synthesizer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Simulacrum Synthesizer — scry 2", Do(Scry{N: 2})),
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b15AnotherBigArtifactYouControlEntered(ev, source, g)
			}, "Simulacrum Synthesizer — create a Construct", Do(CreateToken{Template: UrzaConstructToken(), N: 1})),
		},
	})
}
