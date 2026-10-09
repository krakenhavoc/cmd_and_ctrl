package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Razorfield Ripper — Artifact Creature — Equipment Rhino {2}{W}, 3/3:
//
//	"Whenever this creature or equipped creature attacks, you get {E}
//	 (an energy counter), then it gets +X/+X until end of turn, where X
//	 is the amount of {E} you have.
//	 Reconfigure—Pay {2} or {E}{E}{E}."
//
// The trigger gets the energy first and counts afterwards, so the
// counter it just gave is in X. "It" is the creature that attacked, and
// one that left before the trigger resolved still gets the energy but no
// pump.
//
// "Pay {2} or {E}{E}{E}" is a choice of cost, so each half of reconfigure
// is two rows, one per payment (ReconfigureOneOf, #2639). The energy rows
// pay through the ADR 0129 cost component.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ef842299-8a31-4793-9142-2bcd7a8b2fab",
		Name:         "Razorfield Ripper",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, ThisOrEquippedCreatureAttacks,
				"Razorfield Ripper — you get {E}, then it gets +X/+X", razorfieldRipperEnergyPump),
		},
		Activated: ReconfigureOneOf("Reconfigure—Pay {2} or {E}{E}{E}",
			ReconfigureOption{Pay: "{2}", Cost: ManaCost("{2}")},
			ReconfigureOption{Pay: "{E}{E}{E}", Cost: PayEnergy(3)},
		),
	})
}

func razorfieldRipperEnergyPump(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (GetEnergy{N: 1}).Apply(ctx); err != nil {
		return err
	}
	x := 0
	if p := g.PlayerByIDForEffect(ctx.Controller()); p != nil {
		x = p.Energy
	}
	return pumpTriggeringCreatureUntilEOT(ctx, x, x, "Razorfield Ripper — +X/+X")
}
