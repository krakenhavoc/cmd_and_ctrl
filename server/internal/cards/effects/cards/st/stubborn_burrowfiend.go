package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stubborn Burrowfiend — Creature — Badger Beast Mount {1}{G}:
//
//	"Whenever this creature becomes saddled for the first time each
//	 turn, mill two cards, then this creature gets +X/+X until end of
//	 turn, where X is the number of creature cards in your graveyard.
//	 Saddle 2"
//
// "For the first time each turn" is the event's own shape: the engine
// emits EventBecameSaddled only when the Mount was not already saddled,
// and the designation lasts exactly one turn, so a second saddle in the
// same turn fires nothing. X is counted after the mill, as the printed
// order says, and the pump is read once at resolution (CR 611.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "24bc7bd5-f262-45a2-8e94-377fef2ba5d5",
		Name:         "Stubborn Burrowfiend",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Saddle(2)},
		Triggered: []game.TriggeredAbility{
			// The boost counts the graveyard once the mill's prompt is answered (CR 608.2c, #4155).
			effects.WhenBecomesSaddled("Stubborn Burrowfiend — mill two, then +X/+X", effects.Do(
				effects.MillCards{N: 2},
				effects.Instruction(func(ctx *effects.Context) error {
					if !effects.OnBattlefield(ctx.Game, ctx.Source()) {
						return nil
					}
					x := effects.B11CreatureCardsInGraveyard(ctx.Game, ctx.Controller())
					return effects.BoostUntilEOT{Target: ctx.Source(), Power: x, Toughness: x, Label: "Stubborn Burrowfiend — +X/+X"}.Apply(ctx)
				}),
			)),
		},
	})
}
