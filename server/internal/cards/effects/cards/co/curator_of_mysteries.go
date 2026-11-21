package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Curator of Mysteries — Creature — Sphinx {2}{U}{U}, 4/4:
//
//	"Flying
//	 Whenever you cycle or discard another card, scry 1.
//	 Cycling {U} ({U}, Discard this card: Draw a card.)"
//
// "Cycle or discard" is one event, the discard: cycling's cost is to
// discard the card (CR 702.29d), so watching the cycle event as well would
// count one cycle twice. See Archfiend of Ifnir, which has the same
// trigger; this is its body with a scry.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "04858b93-0a2a-44d0-942d-7fadf589cbb7",
		Name:            "Curator of Mysteries",
		Completeness:    effects.CompletenessFull,
		Purpose:         game.Purpose{Scry: 1}, // the scry half of the discard trigger
		PrintedKeywords: []string{"flying"},
		Activated:       []effects.ActivatedAbility{effects.Cycling("{U}")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDiscardCard, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.DiscardedByYou(ev, source) && ev.CardID != source.InstanceID
			}, "Curator of Mysteries — scry 1", effects.Do(effects.Scry{N: 1})),
		},
	})
}
