package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Baneful Omen — Enchantment {4}{B}{B}{B} (Rise of the Eldrazi, S82 OLD
// batch 481, #5581):
//
//	"At the beginning of your end step, you may reveal the top card of your
//	 library. If you do, each opponent loses life equal to that card's mana
//	 value."
//
// A "you may" trigger (CR 603.5): the reveal is the choice. The mana value
// is read off the revealed card; the loss is life loss, not damage.
//
// No simplification.
func init() {
	t := effects.On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
		return ev.Actor == source.Controller
	}, "Baneful Omen — reveal the top card of your library; each opponent loses life equal to its mana value",
		func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			card, loss, err := effects.RevealTopReadingManaValue(ctx, "Baneful Omen — reveal the top card of your library")
			if err != nil || card == uuid.Nil || loss == 0 {
				return err
			}
			return effects.EachOpponentLosesLifeNow(g, item, loss)
		})
	effects.Register(effects.Spec{
		OracleID:     "dc1711cf-c62b-436b-ba35-56bef8ef14d0",
		Name:         "Baneful Omen",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(t, "Baneful Omen — reveal the top card of your library?"),
		},
	})
}
