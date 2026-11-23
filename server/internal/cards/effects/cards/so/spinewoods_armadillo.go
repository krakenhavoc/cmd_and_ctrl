package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spinewoods Armadillo — Creature — Armadillo {4}{G}{G}, 7/7 (Outlaws of
// Thunder Junction, #4468):
//
//	"Reach
//	 Ward {3} (Whenever this creature becomes the target of a spell or
//	 ability an opponent controls, counter it unless that player pays
//	 {3}.)
//	 {1}{G}, Discard this card: Search your library for a basic land card
//	 or a Desert card, reveal it, put it into your hand, then shuffle.
//	 You gain 3 life."
//
// Ward is the ward.go trigger. The discard ability works only from your
// hand (ADR 0062) and is not cycling, so "whenever you cycle" does not
// see it. Its search is Outcaster Greenblade's predicate, and the life
// is Clay-Fired Bricks's continuation (thenChangeLife): the search may
// stop on a prompt, and "You gain 3 life" is printed after it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8c81f7b6-38e7-482b-afff-727074a23888",
		Name:            "Spinewoods Armadillo",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardMana("{3}"), "Spinewoods Armadillo — ward {3}"),
		},
		Activated: []effects.ActivatedAbility{{
			Label: "{1}{G}, Discard this card: Search your library for a basic land card or a Desert card, reveal it, put it into your hand, then shuffle. You gain 3 life.",
			Cost:  effects.Plus(effects.ManaCost("{1}{G}"), effects.DiscardThis()),
			Zones: []game.ZoneKind{game.ZoneHand},
			// ADR 0142: a tutor and a life gain answer nothing.
			Purpose: game.Purpose{Answers: game.AnswerValue, Tutors: 1},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.SearchLibrary{
					Player:    item.Controller,
					Predicate: effects.B34IsBasicLandOrDesertCard,
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Spinewoods Armadillo — a basic land card or a Desert card",
					ThenTail:  ctx.TailN(effects.ThenChangeLife, 3),
				}.Apply(ctx)
			},
		}},
	})
}
