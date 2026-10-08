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
				WhenThisDies("Absolving Lammasu — you gain 3 life and suspect up to one target creature an opponent controls",
					func(g *game.Game, item *game.StackItem) error {
						if err := (GainLife{Player: item.Controller, Amount: 3}).Apply(NewContext(g, item)); err != nil {
							return err
						}
						return SuspectEachLegalTarget(g, item)
					}),
				UpToOneTargetCreature("up to one target creature an opponent controls", OpponentControls()),
			),
		},
	})
}
