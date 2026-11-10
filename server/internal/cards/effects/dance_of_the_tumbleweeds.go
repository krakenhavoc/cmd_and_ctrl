package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dance of the Tumbleweeds — Sorcery {1}{G} (Outlaws of Thunder Junction):
//
//	"Spree (Choose one or more additional costs.)
//	 + {1} — Search your library for a basic land card or a Desert card, put
//	   it onto the battlefield, then shuffle.
//	 + {3} — Create an X/X green Elemental creature token, where X is the
//	   number of lands you control."
//
// Modes run in printed order, so with both chosen the land the first bullet
// fetches counts for the second. X is read as the second bullet resolves
// (the number of lands you control then), the token is the green Elemental
// row sized to it (Tumbleweed Rising's shape). The search takes a basic land
// card or any card with the subtype Desert; it may fail to find (CR 701.23b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7c57d4aa-d221-4b36-a544-9fa89d63ffbe",
		Name:         "Dance of the Tumbleweeds",
		Completeness: CompletenessFull,
		Modes: Spree(
			ModeWithPurpose(SpreeModeDoing("Search your library for a basic land card or a Desert card, put it onto the battlefield, then shuffle.", "{1}", nil,
				func(item *game.StackItem, ctx *Context, occ int) error {
					return SearchLibrary{
						Player: item.Controller,
						Predicate: func(c game.Card) bool {
							return c.HasSubtype("Desert") || (c.IsLand() && c.HasSupertype("basic"))
						},
						Dest:    game.ZoneBattlefield,
						Limit:   1,
						Reveal:  true,
						Shuffle: true,
						Reason:  "Dance of the Tumbleweeds — a basic land card or a Desert card, onto the battlefield",
					}.Apply(ctx)
				}), game.Purpose{Lands: 1}),
			SpreeModeDoing("Create an X/X green Elemental creature token, where X is the number of lands you control.", "{3}", nil,
				func(item *game.StackItem, ctx *Context, occ int) error {
					x := 0
					for _, c := range ctx.Game.BattlefieldCardsForEffect() {
						if c.Controller == item.Controller && c.IsLand() {
							x++
						}
					}
					t := TokenCard("2/2 green Elemental")
					t.Power, t.Toughness, t.PrintedPTKnown = x, x, true
					return CreateToken{Template: t, N: 1}.Apply(ctx)
				}),
		),
	})
}
