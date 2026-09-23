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
// #1313 / ADR 0058's 2026-09-23 amendment: the activated ability, on
// the same "for as long as this remains tapped" untap hold as Rust
// Tick's.
func init() {
	Register(Spec{
		OracleID:     "1c69fdcf-ba87-480a-88df-70aa4ec9fff0",
		Name:         "Amber Prison",
		Completeness: CompletenessFull,
		UntapOptOuts: []game.UntapOptOut{
			mayChooseNotToUntapSelf("Amber Prison — you may choose not to untap this artifact"),
		},
		Activated: []ActivatedAbility{{
			Label:   "{4}, {T}: Tap target artifact, creature, or land. That permanent doesn't untap during its controller's untap step for as long as this artifact remains tapped.",
			Cost:    Plus(ManaCost("{4}"), TapCost()),
			Targets: TargetPermanent("target artifact, creature, or land", Or(Artifact(), Creature(), Land())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return TapAndHoldWhileThisRemainsTapped(ctx, holdTargetIDs(ctx))
			},
		}},
	})
}
