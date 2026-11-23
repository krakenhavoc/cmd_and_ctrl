package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "facc2d6f-3f28-4b87-b643-44fdf939a2b5",
		Name:            "Mindleech Mass",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer(
				"Mindleech Mass — you may look at that player's hand; if you do, you may cast a spell from among those cards without paying its mana cost",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayChoice{
						Question: "Mindleech Mass — look at that player's hand?",
						YesLabel: "Look",
						NoLabel:  "Don't look",
						Then:     mindleechMassLook,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

var mindleechMassLook = effects.MayYes("confirm/mindleech-mass-look-and-cast", func(ctx *effects.Context, _ game.ConfirmAnswer) error {
	return effects.GapB71LookAtHandThenMayCastFree(ctx.Game, ctx.Item, "Mindleech Mass")
})
