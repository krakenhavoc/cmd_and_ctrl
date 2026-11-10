package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "98441834-7da0-4d9f-88a4-a342a9b2a4c3",
		Name:         "Crawling Sensation",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(AtYourUpkeep("Crawling Sensation — you may mill two cards", Do(MillCards{N: 2})),
				"Crawling Sensation — mill two cards?"),
			TriggersOnlyOnceEachTurn(OncePerBatch(OnAny([]game.EventKind{game.EventZoneMove, game.EventLTB}, gitrogLandToGraveyard,
				"Crawling Sensation — create a 1/1 green Insect creature token",
				Do(CreateToken{Template: TokenCard("1/1 green Insect"), N: 1})))),
		},
	})
}
