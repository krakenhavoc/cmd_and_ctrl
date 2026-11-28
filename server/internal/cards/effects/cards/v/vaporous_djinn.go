package cardsv

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vaporous Djinn — Creature — Djinn {2}{U}{U}, 3/4 (Mirage, S82 OLD b806):
//
//	"Flying
//	 At the beginning of your upkeep, this creature phases out unless you pay
//	 {U}{U}. (While it's phased out, it's treated as though it doesn't exist.
//	 It phases in before you untap during your next untap step.)"
//
// "Unless you pay" is CR 118.12's prompt, asked of the Djinn's controller as
// the trigger resolves; declining, or being unable to pay, phases the Djinn
// out through the phasing primitive (CR 702.26b), which skips a source that
// has left or come back as a new object (CR 400.7). It phases back in during
// its controller's next untap step, before the upkeep, so the tax is asked
// again every turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "66fc64db-7dc4-43fa-8087-281a72286c35",
		Name:            "Vaporous Djinn",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Vaporous Djinn — phases out unless you pay {U}{U}",
				func(g *game.Game, item *game.StackItem) error {
					return effects.PayUnless{
						Chooser:  item.Controller,
						Cost:     "{U}{U}",
						Question: "Vaporous Djinn — pay {U}{U} or phase it out?",
						Else:     vaporousDjinnDeclined,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// vaporousDjinnDeclined is the unpaid branch: the Djinn phases out. An on-disk
// key: never renamed, never reused.
var vaporousDjinnDeclined = effects.PayUnlessElse("pay-unless/s82-old-b806-vaporous-djinn-phase-out",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.PhaseOut{Targets: []uuid.UUID{ctx.Item.SourceCardID}}.Apply(ctx)
	})
