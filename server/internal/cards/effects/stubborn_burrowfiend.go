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
			WhenBecomesSaddled("Stubborn Burrowfiend — mill two, then +X/+X", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if err := (MillCards{Player: item.Controller, N: 2}).Apply(ctx); err != nil {
					return err
				}
				if !onBattlefield(g, item.SourceCardID) {
					return nil
				}
				x := b11CreatureCardsInGraveyard(g, item.Controller)
				return BoostUntilEOT{Target: item.SourceCardID, Power: x, Toughness: x, Label: "Stubborn Burrowfiend — +X/+X"}.Apply(ctx)
			}),
		},
	})
}
