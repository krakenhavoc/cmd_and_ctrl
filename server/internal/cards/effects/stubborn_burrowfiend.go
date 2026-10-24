package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "24bc7bd5-f262-45a2-8e94-377fef2ba5d5",
		Name:         "Stubborn Burrowfiend",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{Saddle(2)},
		Triggered: []game.TriggeredAbility{
			// The boost counts the graveyard once the mill's prompt is answered (CR 608.2c, #4155).
			WhenBecomesSaddled("Stubborn Burrowfiend — mill two, then +X/+X", Do(
				MillCards{N: 2},
				Instruction(func(ctx *Context) error {
					if !onBattlefield(ctx.Game, ctx.Source()) {
						return nil
					}
					x := b11CreatureCardsInGraveyard(ctx.Game, ctx.Controller())
					return BoostUntilEOT{Target: ctx.Source(), Power: x, Toughness: x, Label: "Stubborn Burrowfiend — +X/+X"}.Apply(ctx)
				}),
			)),
		},
	})
}
