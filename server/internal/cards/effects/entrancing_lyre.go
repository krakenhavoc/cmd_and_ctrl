package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Entrancing Lyre — Artifact {3}:
//
//	"You may choose not to untap this artifact during your untap step.
//	 {X}, {T}: Tap target creature with power X or less. It doesn't
//	 untap during its controller's untap step for as long as this
//	 artifact remains tapped."
//
// Rust Tick's shape (the untap-step opt-out, ADR 0070, and the untap
// hold that lasts while this remains tapped, ADR 0058) with ADR 0109
// §9's power bound (#1842): it reads the X announced with the
// activation (CR 602.2b), before the target is chosen, and is
// re-checked as the ability resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9f70b907-586f-4c5d-bb7f-8aadf640ada9",
		Name:         "Entrancing Lyre",
		Completeness: CompletenessFull,
		UntapOptOuts: []game.UntapOptOut{
			mayChooseNotToUntapSelf("Entrancing Lyre — you may choose not to untap this artifact"),
		},
		Activated: []ActivatedAbility{{
			Label:   "{X}, {T}: Tap target creature with power X or less. It doesn't untap during its controller's untap step for as long as this artifact remains tapped.",
			Cost:    Plus(ManaCost("{X}"), TapCost()),
			Targets: TargetCreature("target creature with power X or less").WithPowerAtMostX(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return TapAndHoldWhileThisRemainsTapped(ctx, holdTargetIDs(ctx))
			},
		}},
	})
}
