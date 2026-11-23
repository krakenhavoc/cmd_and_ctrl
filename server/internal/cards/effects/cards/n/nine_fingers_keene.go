package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nine-Fingers Keene — {1}{B}{G}{U} Legendary Creature — Human Rogue 4/4
//
//	"Menace
//	 Ward—Pay 9 life.
//	 Whenever Nine-Fingers Keene deals combat damage to a player, look at the
//	 top nine cards of your library. You may put a Gate card from among them
//	 onto the battlefield. Then if you control nine or more Gates, put the
//	 rest into your hand. Otherwise, put the rest on the bottom of your
//	 library in a random order."
//
// The Gate is an optional "a" pick (a prompt with one answer is still the
// player's), entering as it is put (CR 614). "Nine or more Gates" is read
// after it has entered, counting every Gate the controller controls. A
// library with fewer than nine cards looks at what is there.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e06007cf-dc6e-4b50-93d5-9a7c8ee9c023",
		Name:            "Nine-Fingers Keene",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardLife(9), "Nine-Fingers Keene — ward—pay 9 life"),
			effects.WheneverThisDealsCombatDamageToAPlayer("Nine-Fingers Keene — look at the top nine cards of your library; you may put a Gate card onto the battlefield",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					looked := g.LookAtTopOfLibraryForEffect(ctx.Controller(), 9)
					return effects.PutFromLibraryOntoBattlefield{
						Player:   ctx.Controller(),
						Cards:    looked,
						Match:    effects.HasSubtype("Gate"),
						Max:      1,
						Optional: true,
						Label:    "Nine-Fingers Keene — you may put a Gate card from among them onto the battlefield",
						ThenTail: nineFingersKeeneRest,
						ThenItem: true,
					}.Apply(ctx)
				}),
		},
	})
}

// nineFingersKeeneRest is "Then if you control nine or more Gates, put the
// rest into your hand. Otherwise, put the rest on the bottom of your library
// in a random order." An on-disk key: never renamed, never reused.
var nineFingersKeeneRest = effects.PutFromLibraryTail("nine-fingers-keene/rest",
	func(ctx *effects.Context, res effects.PutFromLibraryResult) error {
		if effects.CountYouControl(ctx, effects.HasSubtype("Gate")) >= 9 {
			ctx.Game.TakeFromLibraryToHandForEffect(res.Player, res.Rest)
			return nil
		}
		return effects.PutRestOnBottomInRandomOrder(ctx.Game, res)
	})
