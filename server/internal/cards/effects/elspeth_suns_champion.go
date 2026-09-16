package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Elspeth, Sun's Champion — Legendary Planeswalker — Elspeth for
// {4}{W}{W}, starting loyalty 4:
//
//	"+1: Create three 1/1 white Soldier creature tokens.
//	 −3: Destroy all creatures with power 4 or greater.
//	 −7: You get an emblem with 'Creatures you control get +2/+2 and
//	     have flying.'"
//
// Both of the abilities anybody activates are wired in full. The
// ultimate is NOT REGISTERED, and that is a deliberate omission
// rather than a stub: emblems have no shape in the engine — they are
// an object in no zone with a continuous effect and no permanent to
// hang it on — and an ability whose label promised an emblem and
// delivered a loyalty payment would be a worse lie than an ability
// that isn't offered. The loyalty still accrues past 7; the day
// emblems land, this card gains one entry and nothing else changes.
//
// The −3 reads CurrentPower, so it catches a 2/2 wearing +1/+1
// counters and an anthem, and spares a 6/6 that something shrank.
// That is the printed card: "power 4 or greater" is the power it has
// when the ability resolves.
func init() {
	Register(Spec{
		OracleID: "05e6b243-48a6-4a42-bc5f-413441de9c33",
		Name:     "Elspeth, Sun's Champion",
		// Printed loyalty reaches the card through deck import
		// (ADR 0032 §1); this is the fallback for tokens, fixtures
		// and the dev spawner.
		StartingLoyalty: 4,
		Activated: []ActivatedAbility{
			{
				Label: "+1: Create three 1/1 white Soldier creature tokens.",
				Cost:  LoyaltyCost(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return CreateToken{
						Controller: item.Controller,
						Template:   TokenCard("1/1 colorless Soldier"),
						N:          3,
					}.Apply(NewContext(g, item))
				},
			},
			{
				Label: "−3: Destroy all creatures with power 4 or greater.",
				Cost:  LoyaltyCost(-3),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					// Snapshot first: destroying moves cards out of
					// the slice this would otherwise be walking, and
					// the power that matters is the power at
					// resolution, before anything has died.
					var doomed []game.Card
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.IsCreature() && c.CurrentPower() >= 4 {
							doomed = append(doomed, c)
						}
					}
					for _, c := range doomed {
						if err := (DestroyTarget{Target: c.InstanceID}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				},
			},
		},
	})
}
