package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pitiless Vizier — {3}{B} Creature — Minotaur Cleric 4/2 (Amonkhet, S82 old
// batch 248):
//
//	"Whenever you cycle or discard a card, this creature gains indestructible
//	 until end of turn."
//
// "Cycle or discard" is ONE event, the discard (Archfiend of Ifnir): cycling's
// cost is to discard the card (CR 702.29d), so watching EventCycle as well
// would count one cycle twice. One trigger per card, as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "340e1078-5158-4199-ab43-976741780b2b",
		Name:         "Pitiless Vizier",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDiscardCard, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.DiscardedByYou(ev, source)
			}, "Pitiless Vizier — it gains indestructible until end of turn",
				effects.ThisGainsKeywordUntilEndOfTurn("indestructible", "Pitiless Vizier — indestructible")),
		},
	})
}
