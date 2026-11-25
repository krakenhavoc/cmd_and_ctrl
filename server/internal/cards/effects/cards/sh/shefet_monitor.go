package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shefet Monitor — {5}{G} Creature — Lizard 6/5 (Amonkhet, S82 old batch 245):
//
//	"Cycling {3}{G}
//	 When you cycle this card, you may search your library for a basic land
//	 card or a Desert card, put it onto the battlefield, then shuffle. (Do this
//	 before you draw.)"
//
// The cycling trigger is fired from the graveyard (CR 702.29c) and goes on the
// stack above the cycling ability, so it resolves before the draw. The search
// is optional, puts the land onto the battlefield untapped and shuffles.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "14f081bb-dd7a-4419-af9b-f75d91e1a6d0",
		Name:         "Shefet Monitor",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Activated:    []effects.ActivatedAbility{effects.Cycling("{3}{G}")},
		Triggered: []game.TriggeredAbility{
			effects.InGraveyard(effects.On(game.EventCycle, effects.Self, "Shefet Monitor — you may search for a basic land card or a Desert card and put it onto the battlefield",
				func(g *game.Game, item *game.StackItem) error {
					return effects.SearchLibrary{
						Player:    item.Controller,
						Source:    item.SourceCardID,
						Predicate: func(c game.Card) bool { return effects.IsBasicLand(c) || c.HasSubtype("Desert") },
						Dest:      game.ZoneBattlefield,
						Limit:     1,
						Shuffle:   true,
						Optional:  true,
						Reason:    "Shefet Monitor — a basic land card or Desert card, onto the battlefield",
					}.Apply(effects.NewContext(g, item))
				})),
		},
	})
}
