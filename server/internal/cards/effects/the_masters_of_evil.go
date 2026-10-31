package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Masters of Evil — {5}{B} Legendary Creature — Human Villain 5/6
// (Marvel Super Heroes, #4455):
//
//	"Other Villains you control get +2/+1.
//	 {1}{B}, Discard this card: Search your library for a Plan card,
//	 reveal it, put it into your hand, then shuffle."
//
// The first line is Attuma's anthem with another tribe and size. The second
// is a hand ability (#660): the discard is a cost, paid at announce, and the
// search is Typecycling's body — reveal, to hand, shuffle even when nothing
// is found (CR 701.23b). It is not cycling (it does not emit a cycle event),
// so it builds its own row. "Plan" is read from the source's text, as
// Attuma's Merfolk is, so a changeling Plan is not a concern but a Plan
// subtype on any card is found.
//
// No simplification.
func init() {
	otherVillains := TribeFilter{Tribes: []string{"Villain"}, Others: true, YoursOnly: true}
	Register(Spec{
		OracleID:     "fcb1caea-e1ac-48f0-a67a-d4aadad490d9",
		Name:         "The Masters of Evil",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(otherVillains, 2, 1),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{B}, Discard this card: Search your library for a Plan card, reveal it, put it into your hand, then shuffle.",
			Cost:    Plus(ManaCost("{1}{B}"), DiscardThis()),
			Zones:   []game.ZoneKind{game.ZoneHand},
			Purpose: game.Purpose{Answers: game.AnswerValue, Tutors: 1},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				plan := ctx.TextSubtype("Plan")
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: func(c game.Card) bool { return c.HasSubtype(plan) },
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "The Masters of Evil — a Plan card, to your hand",
				}.Apply(ctx)
			},
		}},
	})
}
