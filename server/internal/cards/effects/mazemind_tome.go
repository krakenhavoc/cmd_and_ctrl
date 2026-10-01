package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mazemind Tome — Artifact — Book {2}:
//
//	"{T}, Put a page counter on this artifact: Scry 1.
//	 {2}, {T}, Put a page counter on this artifact: Draw a card.
//	 When there are four or more page counters on this artifact, exile
//	 it. If you do, you gain 4 life."
//
// ADR 0107 §1 (#1858). Putting the page counter on is part of each
// row's COST (CR 602.1a), so it is there before the ability resolves and
// the fourth one triggers the exile at once, with the draw or the scry
// still on the stack under it. The exile is a CR 603.8 state trigger, and
// the 4 life comes only if the Tome really reached exile.
//
// No simplification.
func init() {
	const page = "page"
	Register(Spec{
		OracleID:     "800fb917-8898-4eba-9f94-aec6e9d75236",
		Name:         "Mazemind Tome",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:  "{T}, Put a page counter on this artifact: Scry 1.",
				Cost:   Plus(TapCost(), AddCounterToThis(page, 1)),
				Effect: Do(Scry{N: 1}),
			},
			{
				Label:  "{2}, {T}, Put a page counter on this artifact: Draw a card.",
				Cost:   Plus(ManaCost("{2}"), TapCost(), AddCounterToThis(page, 1)),
				Effect: Do(DrawCards{N: 1}),
			},
		},
		Triggered: []game.TriggeredAbility{
			WhenThisHasAtLeast(page, 4, "Mazemind Tome — exile it and gain 4 life",
				ExileThisThen(func(ctx *Context) error {
					return GainLife{Amount: 4}.Apply(ctx)
				})),
		},
	})
}
