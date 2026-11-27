package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thermal Flux — Instant {U} (Coldsnap):
//
//	"Choose one —
//	 • Target nonsnow permanent becomes snow until end of turn.
//	 • Target snow permanent isn't snow until end of turn.
//	 Draw a card at the beginning of the next turn's upkeep."
//
// Each bullet is a layer 4 supertype change (CR 613.1d) lasting to the end of
// the turn, added to or removed from the permanent without touching its types
// (CR 205.4b). The draw follows whichever bullet was chosen, as a delayed
// trigger (CR 603.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "90fe98df-4f87-49ff-b562-0ed51d1d3a1c",
		Name:         "Thermal Flux",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Modes: effects.ChooseOne(
			effects.Mode("Target nonsnow permanent becomes snow until end of turn.",
				effects.TargetPermanent("target nonsnow permanent", effects.Not(effects.Snow()))),
			effects.Mode("Target snow permanent isn't snow until end of turn.",
				effects.TargetPermanent("target snow permanent", effects.Snow())),
		),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if id, ok := effects.B16FirstLegalTargetCard(ctx); ok {
				switch {
				case ctx.HasMode(0):
					if err := effects.UntilEndOfTurn(ctx, id, nil, "Thermal Flux — becomes snow until end of turn", game.AddSupertypesMod("Snow")); err != nil {
						return err
					}
				case ctx.HasMode(1):
					if err := effects.UntilEndOfTurn(ctx, id, nil, "Thermal Flux — isn't snow until end of turn", game.RemoveSupertypesMod("Snow")); err != nil {
						return err
					}
				}
			}
			return effects.DrawACardAtTheNextUpkeep("Thermal Flux — draw a card").Apply(ctx)
		},
	})
}
