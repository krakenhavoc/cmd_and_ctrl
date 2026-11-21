package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deathreap Ritual — Enchantment {2}{B}{G} (EDHREC rank 2066):
//
//	"Morbid — At the beginning of each end step, if a creature died
//	 this turn, you may draw a card."
//
// A card a turn in any game where creatures trade, on anyone's turn.
// The trigger is the any-player end-step condition (b15EndStepBegan)
// with the morbid clause as its intervening if: the tally read
// b11CreaturesDiedThisTurn (every creature that went to a graveyard
// from the battlefield since this turn began, tokens included) is
// checked when the trigger would fire AND again on resolution (CR
// 603.4), and the draw is the trigger's optional prompt.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b26e3596-5b28-4eb6-b3e2-03f63d8c6d49",
		Name:         "Deathreap Ritual",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OptionalDraw(effects.On(game.EventBeginEndStep, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B15EndStepBegan(ev) && effects.B11CreaturesDiedThisTurn(g) > 0
			}, "Deathreap Ritual — draw a card (morbid)", func(g *game.Game, item *game.StackItem) error {
				if effects.B11CreaturesDiedThisTurn(g) == 0 {
					return nil
				}
				return effects.DrawCards{Player: item.Controller, N: 1}.Apply(effects.NewContext(g, item))
			}), "Deathreap Ritual — a creature died this turn: draw a card?", 1),
		},
	})
}
