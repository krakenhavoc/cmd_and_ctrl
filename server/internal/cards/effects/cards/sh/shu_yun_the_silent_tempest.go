package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shu Yun, the Silent Tempest — {2}{U} Legendary Creature — Human Monk 3/2
// (Fate Reforged):
//
//	"Prowess
//	 Whenever you cast a noncreature spell, you may pay {R/W}{R/W}. If you do,
//	 target creature gains double strike until end of turn."
//
// Frenzied Goblin's shape: the target is chosen as the trigger goes on the
// stack (CR 603.3d), before anyone knows whether the mana will be paid; "you
// may pay" is MayPay's prompt as it resolves, and the paid branch grants double
// strike if the target is still legal (CR 608.2b).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7b22f31c-9caa-4afd-8c44-f88346bfa79c",
		Name:            "Shu Yun, the Silent Tempest",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"prowess"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventCast, effects.B10NoncreatureSpellCastByYou, "Shu Yun — you may pay {R/W}{R/W}",
					func(g *game.Game, item *game.StackItem) error {
						return effects.MayPay{
							Chooser:  item.Controller,
							Cost:     "{R/W}{R/W}",
							Question: "Shu Yun — pay {R/W}{R/W} so target creature gains double strike until end of turn?",
							Then:     shuYunPaid,
						}.Apply(effects.NewContext(g, item))
					}),
				effects.TargetCreature("target creature")),
		},
	})
}

// shuYunPaid grants double strike to the chosen creature, if it is still a
// legal target.
var shuYunPaid = effects.PayUnlessBranches("pay-unless/shu-yun-double-strike",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		target, ok := effects.B16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		return effects.GrantKeywordUntilEOT{Target: target, Keywords: []string{"double strike"}, Label: "Shu Yun — double strike until end of turn"}.Apply(ctx)
	}, nil)
