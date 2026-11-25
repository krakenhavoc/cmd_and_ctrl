package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dispossess — {2}{B} Sorcery (Amonkhet, S82 old batch 247):
//
//	"Choose an artifact card name. Search target opponent's graveyard, hand, and
//	 library for any number of cards with the chosen name and exile them. Then
//	 that player shuffles."
//
// The name is chosen as the spell resolves, through the shared card-name
// prompt. The search is the extraction family's (The Stone Brain's): the
// graveyard, hand and library of the targeted opponent in one prompt, any
// number of cards, and the library is shuffled. "An artifact card name" is
// kept where the name is read: only an artifact card of that name is found,
// so naming a card that is not an artifact exiles nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9c2bbab4-97a9-4aad-96ea-31d7ef9e28fe",
		Name:         "Dispossess",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			victim := effects.FirstLegalTargetPlayer(ctx)
			if victim == uuid.Nil {
				return nil
			}
			ctx.Game.QueueCardNameChoiceKeyedForEffect(ctx.Controller(), ctx.Source(),
				"Dispossess — choose an artifact card name",
				effects.WithCarry(ctx.ItemTail(dispossessNamed), victim))
			return nil
		},
	})
}

var dispossessNamed = effects.CardTail("then/s82-old-b247-dispossess-named", func(ctx *effects.Context, o game.TailOutcome) error {
	victim, name := effects.TailCarry0(o), o.Tail.Word
	if victim == uuid.Nil {
		return nil
	}
	return effects.SearchLibrary{
		Player:       ctx.Controller(),
		LibraryOwner: victim,
		Predicate: func(c game.Card) bool {
			return game.CardNameMatches(c, name) && c.IsArtifact()
		},
		Dest:                 game.ZoneExile,
		Unbounded:            true,
		AlsoGraveyardAndHand: true,
		Shuffle:              true,
		Reason:               "Dispossess — search their graveyard, hand and library for any number of cards named " + name,
	}.Apply(ctx)
})
