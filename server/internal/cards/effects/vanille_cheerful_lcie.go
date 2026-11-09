package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vanille, Cheerful l'Cie — Legendary Creature — Human Cleric {3}{G}, 3/2
// (Final Fantasy, #4462):
//
//	"When Vanille enters, mill two cards, then return a permanent card
//	 from your graveyard to your hand.
//	 At the beginning of your first main phase, if you both own and control
//	 Vanille and a creature named Fang, Fearless l'Cie, you may pay
//	 {3}{B}{G}. If you do, exile them, then meld them into Ragnarok,
//	 Divine Deliverance."
//
// The enters trigger is Carrion Cruiser's mill-then-return over permanent
// cards. The meld ability is a beginning-of-main trigger whose condition is
// also an intervening "if" (CR 603.4, ADR 0145): checked as it triggers and
// as it resolves, before the "you may pay" is asked. The paid half is
// MeldWith, which rechecks that both are still owned and controlled. The
// combined back face, Ragnarok, Divine Deliverance, is its own record.
//
// No simplification.
func init() {
	const label = "Vanille, Cheerful l'Cie — you may pay {3}{B}{G}; if you do, exile Vanille and Fang, Fearless l'Cie, then meld them into Ragnarok, Divine Deliverance"
	Register(Spec{
		OracleID:     "577725ae-290f-4081-8352-3317521a096f",
		Name:         "Vanille, Cheerful l'Cie",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Vanille, Cheerful l'Cie — mill two cards, then return a permanent card from your graveyard to your hand",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return MillToZone{N: 2, ThenTail: ctx.ItemTail(vanilleMilled)}.Apply(ctx)
				}),
			InterveningIf(condYouOwnAndControlThisAndANamed{"Fang, Fearless l'Cie", "creature"},
				AtYourPrecombatMain(label, func(g *game.Game, item *game.StackItem) error {
					return MayPay{
						Chooser:  item.Controller,
						Cost:     "{3}{B}{G}",
						Question: "Vanille, Cheerful l'Cie — pay {3}{B}{G} to meld with Fang, Fearless l'Cie?",
						Then:     vanilleMeldPaid,
					}.Apply(NewContext(g, item))
				})),
		},
	})
}
