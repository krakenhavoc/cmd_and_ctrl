package cardsn

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nylea's Colossus — Enchantment Creature — Giant {6}{G}, 6/6 (Commander
// 2018, #5526):
//
//	"Constellation — Whenever this creature or another enchantment you
//	 control enters, double target creature's power and toughness until end
//	 of turn."
//
// Eidolon of Astral Winds' constellation trigger (constellationTriggers),
// targeting any creature as it goes on the stack, with Epic Fight's
// doubling: the creature gets +X/+Y until end of turn, where X and Y are its
// power and toughness as the trigger resolves (CR 701.10b). A target that
// is gone or illegal by then is left alone (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3d73dfd6-fd74-453f-b323-6505825caf4c",
		Name:         "Nylea's Colossus",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventETB, effects.ConstellationTriggers, "Nylea's Colossus — double target creature's power and toughness until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind == game.TargetCard {
							return effects.DoublePowerAndToughnessOfEachUntilEOT(ctx, []uuid.UUID{t.ID}, "Nylea's Colossus — double power and toughness")
						}
					}
					return nil
				}),
				effects.TargetCreature("target creature")),
		},
	})
}
