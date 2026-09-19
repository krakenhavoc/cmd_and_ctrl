package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rust Tick — Artifact Creature — Insect (1/3) for {3}:
//
//	"You may choose not to untap this creature during your untap step.
//	 {1}, {T}: Tap target artifact. It doesn't untap during its
//	 controller's untap step for as long as this creature remains
//	 tapped."
//
// #826 / ADR 0070: the first card on the untap-step OPT-OUT. The first
// clause is the whole of CR 502.3's other half — untapping is
// mandatory unless a card says otherwise, and this is a card saying
// otherwise. A tapped Rust Tick joins the untap step's prompt as a
// candidate its controller may leave out; with nothing else in
// question the prompt's floor is zero, so "untap nothing" is a real
// answer.
//
// Declared simplification: the activated ability is not implemented.
// Its rider is a restriction on a REMEMBERED target that lasts "for as
// long as this creature remains tapped", and the engine has no
// per-source linked-object field — ADR 0058 Decision 8's second
// bullet, still open. ADR 0070 Decision 6 designs it as two more
// fields on ADR 0058's UntapSkip and deliberately does not build it,
// because the tap half without the rider is a materially different
// card and half an ability is worse than none.
func init() {
	Register(Spec{
		OracleID:     "c7f20899-2625-4b3a-8bd8-0bcee07ed86e",
		Name:         "Rust Tick",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Its tap ability isn't implemented — the creature can only be used for its \"you may choose not to untap\" clause.",
		},
		UntapOptOuts: []game.UntapOptOut{
			mayChooseNotToUntapSelf("Rust Tick — you may choose not to untap this creature"),
		},
	})
}
