package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gleam of Resistance — Instant {4}{W} (Conflux, #5587):
//
//	"Creatures you control get +1/+2 until end of turn. Untap those
//	 creatures.
//	 Basic landcycling {1}{W} ({1}{W}, Discard this card: Search your
//	 library for a basic land card, reveal it, put it into your hand,
//	 then shuffle.)"
//
// The team pump of creaturesYouControlGetUntilEOT, then an untap of the
// same creatures: the ones you control as it resolves (CR 611.2c), so a
// creature that arrives later this turn is neither pumped nor untapped.
// Basic landcycling is the shared hand-zone ability (cycling.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c4c0a391-4549-4041-bc6d-09f8b756868e",
		Name:         "Gleam of Resistance",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{BasicLandcycling("{1}{W}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			var those []uuid.UUID
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.IsCreature() && c.Controller == item.Controller {
					those = append(those, c.InstanceID)
				}
			}
			if err := creaturesYouControlGetUntilEOT(1, 2, "Gleam of Resistance — creatures you control get +1/+2")(item, ctx); err != nil {
				return err
			}
			for _, id := range those {
				if err := (UntapTarget{Target: id}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
