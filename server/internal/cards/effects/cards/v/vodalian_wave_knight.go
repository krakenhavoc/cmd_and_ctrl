package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vodalian Wave-Knight — Creature — Merfolk Knight {2}{W}{U}, 3/3 (Commander
// Legends of March of the Machine):
//
//	"Whenever you draw a card, put a +1/+1 counter on each other Merfolk
//	 and/or Knight you control."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "148145ef-a368-4caf-921d-e3809adf03de",
		Name:         "Vodalian Wave-Knight",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor == source.Controller
			}, "Vodalian Wave-Knight — +1/+1 counter on each other Merfolk and/or Knight you control",
				effects.S82Mom09CounterOnEachOtherMerfolkOrKnight),
		},
	})
}
