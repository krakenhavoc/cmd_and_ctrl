package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Insatiable Avarice — Sorcery {B}:
//
//	"Spree (Choose one or more additional costs.)
//	 + {2} — Search your library for a card, then shuffle and put
//	   that card on top.
//	 + {B}{B} — Target player draws three cards and loses 3 life."
//
// Spree proof card #3 (CR 702.172a, ADR 0065's 2026-09-23 amendment):
// one untargeted bullet (a plain one-mana-symbol tutor onto the
// library, ToTop — Vampiric Tutor's real clause) beside one targeted,
// colour-priced bullet — proof that a mode's Cost need not be generic
// mana and that an untargeted bullet composes with a targeted one
// under the same Spree announcement.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ad3e705f-da57-4eff-84d7-2072522de988",
		Name:         "Insatiable Avarice",
		Completeness: CompletenessFull,
		Modes: Spree(
			SpreeModeDoing("Search your library for a card, then shuffle and put that card on top.", "{2}", nil,
				func(item *game.StackItem, ctx *Context, occ int) error {
					return SearchLibrary{
						Player:  item.Controller,
						Dest:    game.ZoneLibrary,
						ToTop:   true,
						Limit:   1,
						Shuffle: true,
						Reason:  "Insatiable Avarice — search your library for a card",
					}.Apply(ctx)
				}),
			SpreeModeDoing("Target player draws three cards and loses 3 life.", "{B}{B}",
				TargetPlayer("target player"),
				func(item *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					if err := (DrawCards{Player: t.ID, N: 3}).Apply(ctx); err != nil {
						return err
					}
					return GainLife{Player: t.ID, Amount: -3}.Apply(ctx)
				}),
		),
	})
}
