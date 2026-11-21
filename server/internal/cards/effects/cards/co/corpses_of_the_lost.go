package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Corpses of the Lost — Enchantment {2}{B} (The Lost Caverns of Ixalan,
// #4470):
//
//	"Skeletons you control get +1/+0 and have haste.
//	 When this enchantment enters, create a 2/2 black Skeleton Pirate
//	 creature token.
//	 At the beginning of your end step, if you descended this turn, you
//	 may pay 1 life. If you do, return this enchantment to its owner's
//	 hand. (You descended if a permanent card was put into your graveyard
//	 from anywhere.)"
//
// The lord half is the tribal builders' anthem and keyword grant
// (tribal.go), over Skeletons you control, the enchantment's own token
// included; "Skeleton" is a word of its text (CR 612.2). The end-step
// trigger is Deep Goblin Skulltaker's intervening if (AtYourEndStepIfDescended,
// CR 603.4, CR 700.11). "You may pay 1 life. If you do" is Seymour Flux's
// payment (MayPay): a player who can't pay life is not asked, and the
// enchantment returns only once the life is paid, and only while it is
// still the permanent the trigger came from (CR 400.7).
//
// No simplification.
func init() {
	skeletons := effects.TribeFilter{Tribes: []string{"Skeleton"}, YoursOnly: true}
	effects.Register(effects.Spec{
		OracleID:     "78bac92f-333f-4c43-bccf-03c3772ee99b",
		Name:         "Corpses of the Lost",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalAnthem(skeletons, 1, 0),
			effects.TribalKeywordGrant(skeletons, "haste"),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Corpses of the Lost — create a 2/2 black Skeleton Pirate creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 black Skeleton Pirate"), N: 1})),
			effects.AtYourEndStepIfDescended("Corpses of the Lost — you may pay 1 life to return it to its owner's hand",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Pay:      effects.LifePayment(1),
						Question: "Corpses of the Lost — pay 1 life to return this enchantment to its owner's hand?",
						Then:     corpsesOfTheLostPaid,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// corpsesOfTheLostPaid is "return this enchantment to its owner's hand".
// An on-disk key: never renamed, never reused.
var corpsesOfTheLostPaid = effects.PayUnlessBranches("pay-unless/corpses-of-the-lost-return",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.ReturnThisPermanentToOwnersHand(ctx.Game, ctx.Item)
	}, nil)
