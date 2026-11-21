package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Yidaro, Wandering Monster — Legendary Creature — Dinosaur Turtle
// {5}{R}{R}, 7/7 (Jumpstart 2022, #5948):
//
//	"Trample, haste
//	 Cycling {1}{R}
//	 When you cycle this card, shuffle it into your library from your
//	 graveyard. If you've cycled a card named Yidaro, Wandering Monster four
//	 or more times this game, put it onto the battlefield from your graveyard
//	 instead. (Do this before you draw.)"
//
// The cycle trigger is watched from the graveyard (CR 702.29c) and goes on the
// stack above the cycling draw, so it resolves before the draw. The count is
// read from the game's event log (one EventCycle per cycling activation, the
// current one included), by the name of the card each one cycled.
//
// No simplification.
func init() {
	const name = "Yidaro, Wandering Monster"
	effects.Register(effects.Spec{
		OracleID:        "6c0e22f2-f0f3-43e6-87c5-c543032112d8",
		Name:            name,
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample", "haste"},
		Activated:       []effects.ActivatedAbility{effects.Cycling("{1}{R}")},
		Triggered: []game.TriggeredAbility{
			effects.InGraveyard(effects.On(game.EventCycle, effects.Self,
				name+" — cycled: shuffle it into your library from your graveyard (or put it onto the battlefield if you've cycled a card named Yidaro four or more times this game)",
				func(g *game.Game, item *game.StackItem) error {
					if !effects.ThisCardStillInTheGraveyard(g, item) {
						return nil
					}
					if effects.GapB159CyclesOfCardNamed(g, item.Controller, name) >= 4 {
						_, err := g.ReturnFromGraveyardWithCountersForEffect(item.SourceCardID, item.Controller, false, nil)
						return err
					}
					return effects.ShuffleThisCardIntoItsOwnersLibrary(g, item)
				})),
		},
	})
}
