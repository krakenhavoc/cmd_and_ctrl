package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nissa's Encouragement — Sorcery {4}{G} (Hour of Devastation, #5534):
//
//	"Search your library and graveyard for a card named Forest, a card named
//	 Brambleweft Behemoth, and a card named Nissa, Genesis Mage. Reveal those
//	 cards, put them into your hand, then shuffle."
//
// Ecological Appreciation's search of library and graveyard together (both
// zones are named, so none is asked and the library is always searched and
// shuffled, CR 701.23a). The three names are distinct, so the different-names
// set rule is "at most one card of each name", which is the clause's "a card
// named" three times; a name with no card in either zone finds none, and the
// rest are still taken (CR 701.23b). The cards found are revealed and put
// into your hand together.
//
// No simplification.
func init() {
	names := map[string]bool{"Forest": true, "Brambleweft Behemoth": true, "Nissa, Genesis Mage": true}
	effects.Register(effects.Spec{
		OracleID:     "55e897e9-5c70-4e32-8f24-3e0fefa319cd",
		Name:         "Nissa's Encouragement",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 3},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return ctx.Game.SearchLibraryThenForEffect(game.SearchLibrarySpec{
				Player:        ctx.Controller(),
				Source:        ctx.Source(),
				Pred:          func(c game.Card) bool { return names[c.Name] },
				Dest:          game.ZoneHand,
				AlsoGraveyard: true,
				BothZones:     true,
				Limit:         3,
				UpTo:          true,
				Reveal:        true,
				Shuffle:       true,
				Rule:          effects.SearchRuleDifferentNames,
				Reason:        "Nissa's Encouragement — a card named Forest, a card named Brambleweft Behemoth, and a card named Nissa, Genesis Mage",
			})
		},
	})
}
