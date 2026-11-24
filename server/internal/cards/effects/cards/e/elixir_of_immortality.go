package cardse

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Elixir of Immortality — Artifact {1} (Secret Lair Drop):
//
//	"{2}, {T}: You gain 5 life. Shuffle this artifact and your graveyard into
//	 their owner's library."
//
// The life comes first. Then the Elixir (if it is still the object that
// activated, CR 400.7) and every card in the controller's graveyard go to
// their owners' libraries as one simultaneous move, and the controller's
// library is shuffled (Finale of Revelation's shuffle, as a tuck's keyed tail).
// A stolen Elixir goes to its owner's library, and that library is shuffled
// too, after the controller's own. The shuffle happens whether or
// not any card moved (CR 701.24c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "204b71fc-4d5f-40dd-903d-56c733171d84",
		Name:         "Elixir of Immortality",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}: You gain 5 life. Shuffle this artifact and your graveyard into their owner's library.",
			Purpose: game.Purpose{LifeGain: 5, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if err := (effects.GainLife{Amount: 5}).Apply(ctx); err != nil {
					return err
				}
				var ids []uuid.UUID
				if p := g.PlayerByIDForEffect(item.Controller); p != nil && p.Graveyard != nil {
					for _, c := range p.Graveyard.Cards {
						ids = append(ids, c.InstanceID)
					}
				}
				var own []uuid.UUID
				owner := item.Controller
				if effects.OnBattlefield(g, item.SourceCardID) && !effects.SourceIsNewObject(g, item) {
					if c, ok := g.LookupCardForEffect(item.SourceCardID); ok && c.Owner != uuid.Nil {
						owner = c.Owner
					}
					own = []uuid.UUID{item.SourceCardID}
				}
				if owner != item.Controller {
					// The Elixir goes to its owner's library, shuffled apart from the controller's.
					if err := g.TuckCardsToLibraryKeyedForEffect(own, game.TuckOptions{}, game.KeyedTail{Ref: effects.ThenShuffle, Player: owner}); err != nil {
						return err
					}
					own = nil
				}
				return g.TuckCardsToLibraryKeyedForEffect(append(ids, own...), game.TuckOptions{},
					game.KeyedTail{Ref: effects.ThenShuffle, Player: item.Controller})
			},
		}},
	})
}
