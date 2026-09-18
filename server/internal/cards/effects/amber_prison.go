package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Amber Prison — Artifact for {4}:
//
//	"You may choose not to untap this artifact during your untap step.
//	 {4}, {T}: Tap target artifact, creature, or land. That permanent
//	 doesn't untap during its controller's untap step for as long as
//	 this artifact remains tapped."
//
// #826 / ADR 0070. Rust Tick's clause on a noncreature permanent, and
// the second card proving the opt-out works for any permanent type:
// the engine asks the question, not the card, and the prompt is the
// same one a Winter Orb cap raises.
//
// Declared simplification: the activated ability is not implemented,
// for the same reason as Rust Tick's — its "for as long as this
// artifact remains tapped" rider needs the per-source linked-object
// field ADR 0058 Decision 8 left open and ADR 0070 Decision 6 designs
// without building.
func init() {
	Register(Spec{
		OracleID:     "1c69fdcf-ba87-480a-88df-70aa4ec9fff0",
		Name:         "Amber Prison",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Its tap ability isn't implemented — the artifact can only be used for its \"you may choose not to untap\" clause.",
		},
		UntapOptOuts: []game.UntapOptOut{
			mayChooseNotToUntapSelf("Amber Prison — you may choose not to untap this artifact"),
		},
	})
}
