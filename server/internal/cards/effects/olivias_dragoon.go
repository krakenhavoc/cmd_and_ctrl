package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Olivia's Dragoon — {1}{B} Creature — Vampire Berserker, 2/2:
//
//	"Discard a card: This creature gains flying until end of turn."
//
// The enabler, and the reason it is in this batch rather than a card
// batch: madness has to work for a discard paid as a COST (CR 601.2h
// / CR 602.2b), not only for one an effect instructs, and the Fiery
// Temper ruling (2022-12-08) says so in as many words. This is the
// cheapest card in Magic that prints such a cost — no mana, no tap,
// just "Discard a card:" — so the cost path is proved end to end by
// activating a 2/2 rather than by a fixture.
//
// The cost is #660's general component (`DiscardACard()`), which pays
// through the one discard helper with `DiscardCauseCost`. That is
// what makes the discard visible to the CR 614 window — and so to
// madness — without the cost site knowing madness exists, and what
// makes it settle without pausing (a cost is one indivisible step, so
// a commander pitched to it goes to the graveyard rather than being
// offered the command zone).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a5019399-91a8-4233-b16f-399718c4be9c",
		Name:         "Olivia's Dragoon",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: "Discard a card: This creature gains flying until end of turn.",
			Cost:  DiscardACard(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return GrantKeywordUntilEOT{
					Target:   item.SourceCardID,
					Keywords: []string{"flying"},
					Label:    "Olivia's Dragoon — flying until end of turn",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
