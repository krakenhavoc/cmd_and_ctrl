package effects

// Harsh Scrutiny — Sorcery {B}:
//
//	"Target opponent reveals their hand. You choose a creature card
//	 from it. That player discards that card. Scry 1."
//
// The revealed-hand pick (ADR 0116) filtered to creature cards, then
// scry 1 (CR 701.22a) on the next line (ADR 0116 §6). The scry is a
// prompt of its own, queued behind the pick, so it is answered after
// the discard as printed (CR 608.2c). You scry even when there is no
// creature card to take (the 2016-09-20 ruling; CR 609.3).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "31338e24-a437-4587-acab-cec46be021e7",
		Name:         "Harsh Scrutiny",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target opponent", Opponent()),
		// The scry waits for the choice from the revealed hand (CR 608.2c, #4155).
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				return ChooseFromRevealedHand{
					Player: TargetedPlayer(ctx),
					Filter: Creature(),
					Label:  "creature card",
				}.Apply(ctx)
			}),
			Scry{N: 1},
		),
	})
}
