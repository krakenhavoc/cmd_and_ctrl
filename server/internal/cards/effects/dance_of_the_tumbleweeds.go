package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dance of the Tumbleweeds — Sorcery {1}{G} (Outlaws of Thunder Junction,
// #4468):
//
//	"Spree (Choose one or more additional costs.)
//	 + {1} — Search your library for a basic land card or a Desert card,
//	   put it onto the battlefield, then shuffle.
//	 + {3} — Create an X/X green Elemental creature token, where X is the
//	   number of lands you control."
//
// Two Spree bullets (CR 702.172a), neither with a target. The search is
// Outcaster Greenblade's (b34IsBasicLandOrDesertCard) sent to the
// battlefield untapped; the land is the player's pick, and they may find
// nothing (CR 701.23b). The bullets resolve in printed order (CR 608.2c),
// so with both chosen the land just found counts toward X, which is read
// as the token is made (Tumbleweed Rising's sized Elemental row).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7c57d4aa-d221-4b36-a544-9fa89d63ffbe",
		Name:         "Dance of the Tumbleweeds",
		Completeness: CompletenessFull,
		Modes: Spree(
			ModeWithPurpose(SpreeModeDoing("Search your library for a basic land card or a Desert card, put it onto the battlefield, then shuffle.", "{1}", nil,
				func(item *game.StackItem, ctx *Context, _ int) error {
					return SearchLibrary{
						Player:    item.Controller,
						Predicate: b34IsBasicLandOrDesertCard,
						Dest:      game.ZoneBattlefield,
						Limit:     1,
						Shuffle:   true,
						Reason:    "Dance of the Tumbleweeds — a basic land card or a Desert card, onto the battlefield",
					}.Apply(ctx)
				}), game.Purpose{Lands: 1}),
			ModeWithPurpose(SpreeModeDoing("Create an X/X green Elemental creature token, where X is the number of lands you control.", "{3}", nil,
				func(item *game.StackItem, ctx *Context, _ int) error {
					x := b03LandsControlled(ctx.Game, item.Controller)
					t := TokenCard("2/2 green Elemental")
					t.Power, t.Toughness, t.PrintedPTKnown = x, x, true
					return CreateToken{Template: t, N: 1}.Apply(ctx)
				}), game.Purpose{Tokens: 1, CreatureTokens: 1}),
		),
	})
}
