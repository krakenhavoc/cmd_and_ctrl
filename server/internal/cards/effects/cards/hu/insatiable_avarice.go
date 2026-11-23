package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "ad3e705f-da57-4eff-84d7-2072522de988",
		Name:         "Insatiable Avarice",
		Completeness: effects.CompletenessFull,
		Modes: effects.Spree(
			effects.SpreeModeDoing("Search your library for a card, then shuffle and put that card on top.", "{2}", nil,
				func(item *game.StackItem, ctx *effects.Context, occ int) error {
					return effects.SearchLibrary{
						Player:  item.Controller,
						Dest:    game.ZoneLibrary,
						ToTop:   true,
						Limit:   1,
						Shuffle: true,
						Reason:  "Insatiable Avarice — search your library for a card",
					}.Apply(ctx)
				}),
			effects.ModeWithPurpose(effects.SpreeModeSteps("Target player draws three cards and loses 3 life.", "{B}{B}",
				effects.TargetPlayer("target player"),
				// The life loss waits for a prompt the draws ask (CR 608.2c, #4155).
				effects.Instruction(func(ctx *effects.Context) error {
					t, ok := effects.ModeTarget(ctx, ctx.Occurrence())
					if !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					return effects.DrawCards{Player: t.ID, N: 3}.Apply(ctx)
				}),
				effects.Instruction(func(ctx *effects.Context) error {
					t, ok := effects.ModeTarget(ctx, ctx.Occurrence())
					if !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					return effects.GainLife{Player: t.ID, Amount: -3}.Apply(ctx)
				})),
				effects.ForTargets(game.TargetPurpose{Slot: 0, Draws: 3, LifeLoss: 3})),
		),
	})
}
