package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crawling Sensation — Enchantment {2}{G} (Tarkir: Dragonstorm
// Commander):
//
//	"At the beginning of your upkeep, you may mill two cards.
//	 Whenever one or more land cards are put into your graveyard from
//	 anywhere for the first time each turn, create a 1/1 green Insect
//	 creature token."
//
// The upkeep half is Hedge Shredder's optional mill. The second is The
// Gitrog Monster's "land cards put into your graveyard from anywhere"
// (both zone-motion events, no OldZone check, a land card you own), made a
// batch trigger (OncePerBatch, CR 603.2c) that fires only the first time in
// a turn (TriggersOnlyOnceEachTurn): lands milled together, or destroyed
// together, are one occurrence, and a later land the same turn makes
// nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "98441834-7da0-4d9f-88a4-a342a9b2a4c3",
		Name:         "Crawling Sensation",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.AtYourUpkeep("Crawling Sensation — you may mill two cards", effects.Do(effects.MillCards{N: 2})),
				"Crawling Sensation — mill two cards?"),
			effects.TriggersOnlyOnceEachTurn(effects.OncePerBatch(effects.OnAny([]game.EventKind{game.EventZoneMove, game.EventLTB, game.EventMill}, effects.GitrogLandToGraveyard,
				"Crawling Sensation — create a 1/1 green Insect creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 green Insect"), N: 1})))),
		},
	})
}
