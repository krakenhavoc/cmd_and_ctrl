package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vastwood Surge — Sorcery {3}{G} (Reprint, #5908):
//
//	"Kicker {4}
//	 Search your library for up to two basic land cards, put them onto the
//	 battlefield tapped, then shuffle. If this spell was kicked, put two
//	 +1/+1 counters on each creature you control."
//
// Explosive Vegetation's search, then the kicked rider. The counters go on
// the creatures you control once the lands have arrived (a land that is a
// creature gets them too).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "8ba2a7ad-a09b-4f52-9c76-c70510dc4feb",
		Name:          "Vastwood Surge",
		Purpose:       game.Purpose{Lands: 2},
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{4}")},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.SearchLibrary{
					Player:        ctx.Controller(),
					Predicate:     effects.IsBasicLand,
					Dest:          game.ZoneBattlefield,
					Limit:         2,
					Reveal:        true,
					Shuffle:       true,
					TappedOnEntry: true,
					Reason:        "Vastwood Surge — up to two basic lands",
				}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				if !ctx.WasKicked() {
					return nil
				}
				return effects.B11PutCountersOnEachCreatureYouControl(ctx.Game, ctx.Item, 2)
			}),
		),
	})
}
