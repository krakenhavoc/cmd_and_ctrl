package cardsn

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Near-Death Experience — Enchantment {2}{W}{W}{W} (Rise of the Eldrazi,
// S82 OLD batch 481, #5581):
//
//	"At the beginning of your upkeep, if you have exactly 1 life, you win
//	 the game."
//
// An intervening "if" (CR 603.4): checked as the upkeep begins and again
// as the trigger resolves.
//
// No simplification.
func init() {
	t := effects.On(game.EventBeginUpkeep, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
		return ev.Actor == source.Controller
	}, "Near-Death Experience — you win the game", func(g *game.Game, item *game.StackItem) error {
		return effects.WinTheGame{Player: item.Controller}.Apply(effects.NewContext(g, item))
	})
	effects.Register(effects.Spec{
		OracleID:     "2afb027f-3c02-46cf-97a1-af0e946ee477",
		Name:         "Near-Death Experience",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIfYou("you have exactly 1 life", nearDeathExperienceExactlyOne, t),
		},
	})
}

// nearDeathExperienceExactlyOne is "you have exactly 1 life".
func nearDeathExperienceExactlyOne(g *game.Game, you uuid.UUID) bool {
	p := g.PlayerByIDForEffect(you)
	return p != nil && p.Life == 1
}
