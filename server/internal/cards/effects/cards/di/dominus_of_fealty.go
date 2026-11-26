package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dominus of Fealty — Creature — Spirit Avatar {U/R}{U/R}{U/R}{U/R}{U/R}, 4/4
// (Commander 2011, #5574):
//
//	"Flying
//	 At the beginning of your upkeep, you may gain control of target permanent
//	 until end of turn. If you do, untap it and it gains haste until end of
//	 turn."
//
// An optional, targeted upkeep trigger over any permanent (the target is chosen
// as it goes on the stack, the "you may" asked after it, CR 603.5): the steal,
// the untap and the haste are Threaten's single effect (StdB31Threaten), in
// printed order. A permanent that is not a creature simply has no use for haste.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f5e18b27-c5cc-4f5e-b767-4b1b43e00fd0",
		Name:            "Dominus of Fealty",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(effects.AtYourUpkeep("Dominus of Fealty — gain control of target permanent until end of turn, untap it, it gains haste",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					t, ok := effects.FirstLegalTarget(ctx)
					if !ok {
						return nil
					}
					return effects.StdB31Threaten(ctx, t, "Dominus of Fealty")
				}),
				effects.TargetPermanent("target permanent")),
				"Dominus of Fealty — gain control of target permanent until end of turn?"),
		},
	})
}
