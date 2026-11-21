package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Troop of Ponies — Creature — Horse {2} (The Hobbit), 2/1:
//
//	"{2}, {T}, Sacrifice this creature: Search your library for up to two
//	 basic land cards, reveal them, put one onto the battlefield tapped and
//	 the other into your hand, then shuffle."
//
// #4454, S76. Cultivate's two searches (the battlefield half chains the hand
// half through a registered tail), as an activated ability whose cost is paid
// up front: the ponies are already in the graveyard when the searches run.
// The hand half's key is this card's own, since a tail's prompt names it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bb52b3c3-09f0-4506-b7c0-250ab0b91866",
		Name:         "Troop of Ponies",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}, Sacrifice this creature: Search your library for up to two basic land cards, reveal them, put one onto the battlefield tapped and the other into your hand, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Lands: 2, Tutors: 1},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost(), effects.SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				controller, source := item.Controller, item.SourceCardID
				return effects.SearchLibrary{
					Player:        controller,
					Source:        source,
					Predicate:     effects.IsBasicLand,
					Dest:          game.ZoneBattlefield,
					Limit:         1,
					Reveal:        true,
					TappedOnEntry: true,
					Reason:        "Troop of Ponies — basic land onto the battlefield tapped",
					ThenTail:      game.KeyedTail{Ref: hobB5PoniesHandHalf, Player: controller, Source: source},
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}

// hobB5PoniesHandHalf is the second search: another basic land into your
// hand, then the shuffle. An on-disk key.
var hobB5PoniesHandHalf = effects.CultivateHandTail("troop-of-ponies/basic-land-to-hand", "Troop of Ponies")
