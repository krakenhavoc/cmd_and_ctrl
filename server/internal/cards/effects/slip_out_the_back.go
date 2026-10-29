package effects

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
		OnResolve: ResolveSteps(
			// The phasing waits for a prompt the counter asks (CR 608.2c, #4155).
			Instruction(putPlusOneCounterOnTheLegalTarget),
			Instruction(func(ctx *Context) error {
				targets := legalTargetCards(ctx.Item, ctx.Game)
				if len(targets) == 0 {
					return nil
				}
				return PhaseOut{Targets: targets}.Apply(ctx)
			}),
		),
	})
}
