package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Start Your Engines — Sorcery {3}{R} (Kaladesh, #5539):
//
//	"Vehicles you control become artifact creatures until end of turn.
//	 Creatures you control get +2/+0 until end of turn."
//
// Nothing targets. The Vehicles are animated first, as Armed and Armored's
// are (BecomeCreatureUntilEOT keeps the printed power and toughness), and
// the +2/+0 then goes to the creatures you control as the spell resolves,
// the freshly animated Vehicles included (CR 611.2c): one that enters later
// this turn is not pumped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bbe3726b-e14d-48d1-8bbf-40d48f1ef8c7",
		Name:         "Start Your Engines",
		Completeness: effects.CompletenessFull,
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				you := ctx.Controller()
				vehicle := ctx.TextSubtype("Vehicle")
				for _, c := range ctx.Game.BattlefieldCardsForEffect() {
					if c.Controller != you || !c.HasSubtype(vehicle) {
						continue
					}
					if err := (effects.BecomeCreatureUntilEOT{Target: c.InstanceID, Label: "Start Your Engines — becomes an artifact creature"}).Apply(ctx.AsGroupMember()); err != nil {
						return err
					}
				}
				return nil
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.BoostUntilEOT{
					Match: effects.And(effects.Creature(), effects.YouControl()), Power: 2,
					Label: "Start Your Engines — +2/+0",
				}.Apply(ctx)
			}),
		),
	})
}
