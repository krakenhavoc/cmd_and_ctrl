package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	skeletons := TribeFilter{Tribes: []string{"Skeleton"}, YoursOnly: true}
	Register(Spec{
		OracleID:     "78bac92f-333f-4c43-bccf-03c3772ee99b",
		Name:         "Corpses of the Lost",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(skeletons, 1, 0),
			TribalKeywordGrant(skeletons, "haste"),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Corpses of the Lost — create a 2/2 black Skeleton Pirate creature token",
				Do(CreateToken{Template: TokenCard("2/2 black Skeleton Pirate"), N: 1})),
			AtYourEndStepIfDescended("Corpses of the Lost — you may pay 1 life to return it to its owner's hand",
				func(g *game.Game, item *game.StackItem) error {
					return MayPay{
						Chooser:  item.Controller,
						Pay:      LifePayment(1),
						Question: "Corpses of the Lost — pay 1 life to return this enchantment to its owner's hand?",
						Then:     corpsesOfTheLostPaid,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

// corpsesOfTheLostPaid is "return this enchantment to its owner's hand".
// An on-disk key: never renamed, never reused.
var corpsesOfTheLostPaid = PayUnlessBranches("pay-unless/corpses-of-the-lost-return",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return returnThisPermanentToOwnersHand(ctx.Game, ctx.Item)
	}, nil)
