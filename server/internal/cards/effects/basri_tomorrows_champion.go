package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "d8e1e9ba-708c-4f54-abc3-817004a2f2ac",
		Name:         "Basri, Tomorrow's Champion",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{W}, {T}, Exert Basri: Create a 1/1 white Cat creature token with lifelink.",
				Cost:    Plus(ManaCost("{W}"), TapCost(), ExertThis()),
				Purpose: game.Purpose{Tokens: 1},
				Effect: func(g *game.Game, item *game.StackItem) error {
					return CreateToken{Template: TokenCard("1/1 white Cat with lifelink"), N: 1}.Apply(NewContext(g, item))
				},
			},
			Cycling("{2}{W}"),
		},
		Triggered: []game.TriggeredAbility{
			InGraveyard(On(game.EventCycle, Self,
				"Basri, Tomorrow's Champion — cycled: Cats you control gain hexproof and indestructible",
				func(g *game.Game, item *game.StackItem) error {
					return GrantKeywordUntilEOT{
						Match:    And(OfCreatureType("Cat"), YouControl()),
						Keywords: []string{"hexproof", "indestructible"},
						Label:    "Basri, Tomorrow's Champion — hexproof and indestructible",
					}.Apply(NewContext(g, item))
				})),
		},
	})
}
