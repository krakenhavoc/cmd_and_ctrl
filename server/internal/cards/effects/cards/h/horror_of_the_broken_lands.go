package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Horror of the Broken Lands — {4}{B} Creature — Horror 4/4 (Amonkhet, S82 old
// batch 245):
//
//	"Whenever you cycle or discard another card, this creature gets +2/+1
//	 until end of turn.
//	 Cycling {B}"
//
// "Cycle or discard" is ONE event, the discard (Archfiend of Ifnir): cycling's
// cost is to discard the card (CR 702.29d), so a cycled card is a discarded
// card and watching EventCycle as well would count one cycle twice. "Another"
// rules out this creature's own cycling.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6de20c2c-b8a8-4f12-897b-a2f55d7eed87",
		Name:         "Horror of the Broken Lands",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Cycling("{B}")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDiscardCard, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.DiscardedByYou(ev, source) && ev.CardID != source.InstanceID
			}, "Horror of the Broken Lands — it gets +2/+1 until end of turn",
				effects.ThisGetsUntilEndOfTurn(2, 1, "Horror of the Broken Lands — +2/+1")),
		},
	})
}
