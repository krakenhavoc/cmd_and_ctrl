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
// #1313 / ADR 0058's 2026-09-23 amendment: the activated ability. Its
// rider is an untap HOLD with the WhileSourceRemainsTapped duration
// ADR 0070 Decision 6 designed. The two clauses are the card's whole
// point together: decline to untap Rust Tick and the artifact stays
// locked; untap it and the lock is over for good.
func init() {
	Register(Spec{
		OracleID:     "c7f20899-2625-4b3a-8bd8-0bcee07ed86e",
		Name:         "Rust Tick",
		Completeness: CompletenessFull,
		UntapOptOuts: []game.UntapOptOut{
			mayChooseNotToUntapSelf("Rust Tick — you may choose not to untap this creature"),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}: Tap target artifact. It doesn't untap during its controller's untap step for as long as this creature remains tapped.",
			Cost:    Plus(ManaCost("{1}"), TapCost()),
			Targets: TargetPermanent("target artifact", Artifact()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return TapAndHoldWhileThisRemainsTapped(ctx, holdTargetIDs(ctx))
			},
		}},
	})
}
