package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Slip Out the Back — Instant {U}:
//
//	"Put a +1/+1 counter on target creature. It phases out. (Treat it
//	 and anything attached to it as though they don't exist until its
//	 controller's next turn.)"
//
// A one-mana save for a voltron creature: the counter goes on first,
// then the creature phases out carrying it, together with every Aura
// and Equipment attached to it (CR 702.26g). It comes back during its
// controller's next untap step with the counter, the attachments and
// its damage intact (CR 702.26d), because phasing is not a zone change.
//
// "It" is the same target, so the phase-out reads the same still-legal
// target the counter went on (CR 608.2b); a creature that left in
// response gets neither half.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "82506e51-3fe1-446f-83c9-69e67583cefc",
		Name:         "Slip Out the Back",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			targets := legalTargetCards(item, ctx.Game)
			if len(targets) == 0 {
				return nil
			}
			if err := (AddCounter{Target: targets[0], Kind: "+1/+1", N: 1}).Apply(ctx); err != nil {
				return err
			}
			return PhaseOut{Targets: targets}.Apply(ctx)
		},
	})
}
