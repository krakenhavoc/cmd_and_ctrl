package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:      "8ba2a7ad-a09b-4f52-9c76-c70510dc4feb",
		Name:          "Vastwood Surge",
		Purpose:       game.Purpose{Lands: 2},
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Kicker("{4}")},
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				return SearchLibrary{
					Player:        ctx.Controller(),
					Predicate:     IsBasicLand,
					Dest:          game.ZoneBattlefield,
					Limit:         2,
					Reveal:        true,
					Shuffle:       true,
					TappedOnEntry: true,
					Reason:        "Vastwood Surge — up to two basic lands",
				}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error {
				if !ctx.WasKicked() {
					return nil
				}
				return b11PutCountersOnEachCreatureYouControl(ctx.Game, ctx.Item, 2)
			}),
		),
	})
}
