package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Absolving Lammasu — Creature — Lammasu {4}{W}:
//
//	"Flying
//	 When this creature enters, all suspected creatures are no longer
//	 suspected.
//	 When this creature dies, you gain 3 life and suspect up to one
//	 target creature an opponent controls. (A suspected creature has
//	 menace and can't block.)"
//
// "All suspected creatures" is every suspected creature on the
// battlefield, whoever controls it (UnsuspectAll with no Match). The
// dies trigger gains the life first and suspects whatever is still a
// legal target when it resolves; with no target chosen it only gains.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "dbedd45e-f76c-4be7-ad89-ff57bed6626f",
		Name:            "Absolving Lammasu",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Absolving Lammasu — all suspected creatures are no longer suspected",
				func(g *game.Game, item *game.StackItem) error {
					return UnsuspectAll{}.Apply(NewContext(g, item))
				}),
			Targeting(
				// The suspect waits for a prompt the life gain asks (a CR 616 order; CR 608.2c, #4155).
				WhenThisDies("Absolving Lammasu — you gain 3 life and suspect up to one target creature an opponent controls",
					Do(
						GainLife{Amount: 3},
						Instruction(func(ctx *Context) error {
							return SuspectEachLegalTarget(ctx.Game, ctx.Item)
						}),
					)),
				UpToOneTargetCreature("up to one target creature an opponent controls", OpponentControls()),
			),
		},
	})
}
