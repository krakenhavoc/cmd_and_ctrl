package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tiger-Seal — Creature — Cat Seal {U}, 3/3 (Avatar: The Last Airbender,
// #4459):
//
//	"Vigilance
//	 At the beginning of your upkeep, tap this creature.
//	 Whenever you draw your second card each turn, untap this creature."
//
// Two printed triggers, two rows. The second-card trigger is Knights of
// Dol Amroth's (b41DrewYourSecondCardThisTurn): once a turn at most, on
// the second draw only, and not at all if the Seal wasn't there for it
// (the rulings). Several cards drawn by one effect trigger it after the
// effect finishes (the ruling): the trigger waits on the stack like any
// other. Neither does anything to a new object once this one has left
// (TapTarget's and UntapTarget's guards, CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "68f84643-c0f9-4d31-ba2e-38b65ab45d04",
		Name:            "Tiger-Seal",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Tiger-Seal — tap this creature", func(g *game.Game, item *game.StackItem) error {
				return effects.TapTarget{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
			}),
			effects.On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B41DrewYourSecondCardThisTurn(ev, source, g)
			}, "Tiger-Seal — untap this creature", func(g *game.Game, item *game.StackItem) error {
				return effects.UntapTarget{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
