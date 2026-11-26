package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mayael the Anima — Legendary Creature — Elf Shaman {R}{G}{W}, 2/3 (Commander
// 2013, S82 OLD b385):
//
//	"{3}{R}{G}{W}, {T}: Look at the top five cards of your library. You may put
//	 a creature card with power 5 or greater from among them onto the
//	 battlefield. Put the rest on the bottom of your library in any order."
//
// Muzzio's look-and-put: the five cards are looked at (not revealed), one
// creature card whose printed power is 5 or greater may be put onto the
// battlefield, and the rest go to the bottom in the order the player picks.
// The power is read on the card in the library (CR 208.3: a card with no
// power, or */*, is not a creature card with power 5 or greater).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7f546d54-584d-4bec-8fbb-1ea2f8ab277e",
		Name:         "Mayael the Anima",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{R}{G}{W}, {T}: Look at the top five cards of your library. You may put a creature card with power 5 or greater from among them onto the battlefield. Put the rest on the bottom of your library in any order.",
			Cost:    effects.Plus(effects.ManaCost("{3}{R}{G}{W}"), effects.TapCost()),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				cards := g.LookAtTopOfLibraryForEffect(ctx.Controller(), 5)
				if len(cards) == 0 {
					return nil
				}
				return effects.PutFromLibraryOntoBattlefield{
					Player:   ctx.Controller(),
					Cards:    cards,
					Match:    effects.And(effects.Creature(), effects.PowerGE(5)),
					Max:      1,
					Optional: true,
					Label:    "Mayael the Anima — you may put a creature card with power 5 or greater onto the battlefield",
					ThenTail: effects.RprB15RestOnBottomAnyOrderTail,
				}.Apply(ctx)
			},
		}},
	})
}
