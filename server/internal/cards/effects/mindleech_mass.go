package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mindleech Mass — Creature — Horror {5}{U}{B}{B}, 6/6 (Ravnica Remastered,
// #5948):
//
//	"Trample
//	 Whenever this creature deals combat damage to a player, you may look at
//	 that player's hand. If you do, you may cast a spell from among those
//	 cards without paying its mana cost."
//
// Silent-Blade Oni's trigger behind a yes/no for the look: on a yes the hand
// is revealed to the Mass's controller alone (LookAtHandForEffect) and they
// are offered the spells among those cards that can be cast free. The card
// is cast from the other player's hand (CR 601.2); a hand with nothing castable
// asks nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "facc2d6f-3f28-4b87-b643-44fdf939a2b5",
		Name:            "Mindleech Mass",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer(
				"Mindleech Mass — you may look at that player's hand; if you do, you may cast a spell from among those cards without paying its mana cost",
				func(g *game.Game, item *game.StackItem) error {
					return MayChoice{
						Question: "Mindleech Mass — look at that player's hand?",
						YesLabel: "Look",
						NoLabel:  "Don't look",
						Then:     mindleechMassLook,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

var mindleechMassLook = MayYes("confirm/mindleech-mass-look-and-cast", func(ctx *Context, _ game.ConfirmAnswer) error {
	return gapB71LookAtHandThenMayCastFree(ctx.Game, ctx.Item, "Mindleech Mass")
})
