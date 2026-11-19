package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eye of Ugin — Legendary Land (Special Guests, S82 BNS):
//
//	"Colorless Eldrazi spells you cast cost {2} less to cast.
//	 {7}, {T}: Search your library for a colorless creature card, reveal it,
//	 put it into your hand, then shuffle."
//
// The discount is generic-only (CostsLess), read off the spell being cast:
// an Eldrazi that has any colour is not discounted, and neither is a
// colorless spell that is not an Eldrazi. The search is Sanctum of Ugin's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "10d13ff6-c4d0-4753-8939-a8a90f0e92bb",
		Name:         "Eye of Ugin",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsLess(2, "Colorless Eldrazi spells you cast cost {2} less to cast.",
				YourSpell(), func(q game.CostQuery) bool {
					return q.Card.IsColorless() && q.Card.HasSubtype(q.Source.TextSubtype("Eldrazi"))
				}),
		},
		Activated: []ActivatedAbility{{
			Label:   "{7}, {T}: Search your library for a colorless creature card, reveal it, put it into your hand, then shuffle.",
			Purpose: game.Purpose{Tutors: 1, Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{7}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: func(c game.Card) bool { return c.IsCreature() && c.IsColorless() },
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Eye of Ugin — a colorless creature card",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
