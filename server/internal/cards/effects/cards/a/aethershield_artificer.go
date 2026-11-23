package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Aethershield Artificer — Creature — Dwarf Artificer {3}{W}, 3/3 (Jumpstart 2022, #5948):
//
//	"At the beginning of combat on your turn, target artifact creature you
//	 control gets +2/+2 and gains indestructible until end of turn."
//
// The beginning-of-combat row (it targets as it goes on the stack, and is
// removed if no artifact creature is there to target, CR 603.3d); the pump
// and the keyword are the two shared until-end-of-turn effects on the same
// target, applied together as the ability resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d2575513-72fa-49c7-be6b-9f6f85950f88",
		Name:         "Aethershield Artificer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.AtBeginningOfYourCombat("Aethershield Artificer — target artifact creature you control gets +2/+2 and gains indestructible until end of turn",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						id, ok := effects.B16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						if err := (effects.BoostUntilEOT{Target: id, Power: 2, Toughness: 2, Label: "Aethershield Artificer — +2/+2"}).Apply(ctx); err != nil {
							return err
						}
						return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"indestructible"}, Label: "Aethershield Artificer — indestructible"}.Apply(ctx)
					}),
				effects.TargetCreature("target artifact creature you control", effects.YouControl(), effects.Artifact())),
		},
	})
}
