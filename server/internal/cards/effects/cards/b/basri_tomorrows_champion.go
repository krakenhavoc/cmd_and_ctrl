package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Basri, Tomorrow's Champion — Legendary Creature — Human Knight {W}, 2/1:
//
//	"{W}, {T}, Exert Basri: Create a 1/1 white Cat creature token with
//	 lifelink. (An exerted creature won't untap during your next untap
//	 step.)
//	 Cycling {2}{W} ({2}{W}, Discard this card: Draw a card.)
//	 When you cycle this card, Cats you control gain hexproof and
//	 indestructible until end of turn."
//
// The activation pays {W}, {T} and an exert (ADR 0130 §4, CR 701.43a).
// The cycle trigger watches EventCycle from the graveyard, where the
// card is once it has been cycled, and grants the keywords to the Cats
// you control as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d8e1e9ba-708c-4f54-abc3-817004a2f2ac",
		Name:         "Basri, Tomorrow's Champion",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{W}, {T}, Exert Basri: Create a 1/1 white Cat creature token with lifelink.",
				Cost:    effects.Plus(effects.ManaCost("{W}"), effects.TapCost(), effects.ExertThis()),
				Purpose: game.Purpose{Answers: game.AnswerMakesBlocker, Tokens: 1},
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.CreateToken{Template: effects.TokenCard("1/1 white Cat with lifelink"), N: 1}.Apply(effects.NewContext(g, item))
				},
			},
			effects.Cycling("{2}{W}"),
		},
		Triggered: []game.TriggeredAbility{
			effects.InGraveyard(effects.On(game.EventCycle, effects.Self,
				"Basri, Tomorrow's Champion — cycled: Cats you control gain hexproof and indestructible",
				func(g *game.Game, item *game.StackItem) error {
					return effects.GrantKeywordUntilEOT{
						Match:    effects.And(effects.OfCreatureType("Cat"), effects.YouControl()),
						Keywords: []string{"hexproof", "indestructible"},
						Label:    "Basri, Tomorrow's Champion — hexproof and indestructible",
					}.Apply(effects.NewContext(g, item))
				})),
		},
	})
}
