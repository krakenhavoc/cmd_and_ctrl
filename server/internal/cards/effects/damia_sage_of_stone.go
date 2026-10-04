package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Damia, Sage of Stone — Legendary Creature — Gorgon Wizard {4}{B}{G}{U},
// 4/4:
//
//	"Deathtouch
//	 Skip your draw step.
//	 At the beginning of your upkeep, if you have fewer than seven
//	 cards in hand, draw cards equal to the difference."
//
// "Skip your draw step" is Necropotence's SkipYourDrawStep
// (CR 614.10 / CR 500.11), declared on Spec.Replacements so it lasts
// exactly as long as Damia is on the battlefield.
//
// The upkeep ability has an intervening "if" (CR 603.4), so the hand
// size is asked twice: when the upkeep begins (the trigger exists only
// with fewer than seven cards) and again as the ability resolves. The
// second look is free, because the effect IS a read of the hand: it
// draws seven minus the hand size now, and nothing when that is not
// positive. A card cast or discarded in response changes the number
// drawn, exactly as the difference is "equal to" something measured on
// resolution.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "eb7ae3e8-0489-4f3a-8343-7e7ec2fe3f01",
		Name:            "Damia, Sage of Stone",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Replacements:    []game.ReplacementEffect{SkipYourDrawStep()},
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginUpkeep,
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return ev.Actor == source.Controller && damiaCardsShort(g, source.Controller) > 0
				},
				"Damia, Sage of Stone — draw up to seven cards",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return DrawCards{Player: item.Controller, N: damiaCardsShort(g, item.Controller)}.Apply(ctx)
				}),
		},
	})
}

// damiaCardsShort is seven minus the cards in `player`'s hand: how far
// below seven it is, or zero.
func damiaCardsShort(g *game.Game, player uuid.UUID) int {
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Hand == nil {
		return 0
	}
	if n := 7 - p.Hand.Size(); n > 0 {
		return n
	}
	return 0
}
