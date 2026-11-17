package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Herd Migration — Sorcery {6}{G}:
//
//	"Domain — Create a 3/3 green Beast creature token for each basic land
//	 type among lands you control.
//	 {1}{G}, Discard this card: Search your library for a basic land card,
//	 reveal it, put it into your hand, then shuffle. You gain 3 life."
//
// The spell is Ordered Migration's count of tokens (DomainCount, #5666)
// with the shared "3/3 green Beast" row. The ability works from the hand
// (Altanak's discard-self cost, CR 602.2): the card is discarded as a cost,
// so it cannot be countered, and the search, the shuffle and the three life
// are one resolution, the life gained even when the search finds nothing.
//
// No simplification.
func init() {
	const ability = "{1}{G}, Discard this card: Search your library for a basic land card, reveal it, put it into your hand, then shuffle. You gain 3 life."
	Register(Spec{
		OracleID:     "dfdc0eab-309b-4441-a94c-930c6dde4c48",
		Name:         "Herd Migration",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			n := DomainCount(ctx.Game, ctx.Controller())
			if n == 0 {
				return nil
			}
			return CreateToken{Controller: ctx.Controller(), Template: TokenCard("3/3 green Beast"), N: n}.Apply(ctx)
		},
		Activated: []ActivatedAbility{{
			Label:   ability,
			Purpose: game.Purpose{Lands: 1, LifeGain: 3, Answers: game.AnswerValue},
			Cost:    game.AbilityCost{Mana: "{1}{G}", DiscardSelf: true},
			Zones:   []game.ZoneKind{game.ZoneHand},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if err := (SearchLibrary{
					Player:    item.Controller,
					Predicate: IsBasicLand,
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Herd Migration — a basic land card, into your hand",
				}).Apply(ctx); err != nil {
					return err
				}
				return GainLife{Amount: 3}.Apply(ctx)
			},
		}},
	})
}
