package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sygg, River Cutthroat — Legendary Creature — Merfolk Rogue
// {U/B}{U/B}, 1/3 (EDHREC rank 3821):
//
//	"At the beginning of each end step, if an opponent lost 3 or more
//	 life this turn, you may draw a card. (Damage causes loss of
//	 life.)"
//
// A card a turn for anyone who bleeds. The trigger watches EVERY
// player's end step; the intervening if (CR 603.4) is checked at
// announce — the trigger does not fire at all when no single
// opponent has lost 3 this turn — and again at resolution, off the
// per-turn tally
// (b06AnOpponentLostAtLeastThisTurn: damage and life loss both,
// per opponent, not summed across the table). The "you may" is the
// trigger's yes/no prompt.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cd4db500-0017-46c6-be94-1bf48f686b6a",
		Name:         "Sygg, River Cutthroat",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OptionalDraw(effects.On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B36EndStepAndAnOpponentLostThree(ev, source, g)
			}, "Sygg, River Cutthroat — draw a card", effects.B36DrawIfAnOpponentLostThree), "Sygg, River Cutthroat — an opponent lost 3 or more life this turn. Draw a card?", 1),
		},
	})
}
