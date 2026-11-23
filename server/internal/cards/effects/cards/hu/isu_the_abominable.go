package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Isu the Abominable — Legendary Snow Creature — Yeti {3}{U}{U}, 5/5
// (Jumpstart 2022, #5948):
//
//	"You may look at the top card of your library any time.
//	 You may play snow lands and cast snow spells from the top of your
//	 library.
//	 Whenever another snow permanent you control enters, you may pay {G},
//	 {W}, or {U}. If you do, put a +1/+1 counter on Isu."
//
// The look is the private strength (LibraryTopOwner) and the permission is
// Courser of Kruphix's and Vizier of the Menagerie's, narrowed to the snow
// supertype (PermissionFilter.SnowOnly): a snow land is played and a snow
// spell cast, at its printed cost, from the top. The trigger watches any other
// permanent, token or card, with the snow supertype entering under your
// control. "Pay {G}, {W}, or {U}" is one mana of any of the three, a single
// three-way hybrid payment; declining or being unable to pay puts no counter.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:          "b34823a3-7bee-4e71-a3df-9b533c8ddf15",
		Name:              "Isu the Abominable",
		Completeness:      effects.CompletenessFull,
		LibraryTopVisible: game.LibraryTopOwner,
		CastPermissions: []game.CastPermission{
			effects.PlayFromTopOfYourLibrary(game.PermissionFilter{SnowOnly: true},
				"Play a snow land or cast a snow spell from the top of your library (Isu the Abominable)"),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.GapB154AnotherSnowPermanentEntersUnderYourControl,
				"Isu the Abominable — you may pay {G}, {W}, or {U}; if you do, put a +1/+1 counter on it",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.MayPay{
						Chooser:  ctx.Controller(),
						Cost:     effects.GapB154GreatSnowPay,
						Question: "Isu the Abominable — pay {G}, {W}, or {U} to put a +1/+1 counter on it?",
						Then:     isuPaid,
					}.Apply(ctx)
				}),
		},
	})
}

// isuPaid is the paid half: a +1/+1 counter on Isu, while it is still on the
// battlefield.
var isuPaid = effects.PayUnlessBranches("pay-unless/s82-gap-b154-isu-counter",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		if !effects.OnBattlefield(ctx.Game, ctx.Source()) {
			return nil
		}
		return effects.AddCounter{Target: ctx.Source(), Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
	}, nil)
