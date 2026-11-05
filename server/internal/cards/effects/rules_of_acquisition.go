package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rules of Acquisition — Sorcery {3}{U}{U} (Star Trek, #4973):
//
//	"Choose one —
//	 • Follow the 3rd Rule — Gain control of target nonland permanent an
//	   opponent controls. That opponent draws three cards.
//	 • Follow the 98th Rule — Target opponent gains control of target
//	   nonland permanent you control. If they do, you draw three cards."
//
// Both control changes are layer-2 effects with no stated duration, so
// they last for the rest of the game (CR 613.1b, CR 611.2a). "That
// opponent" is the player who controlled the permanent as the spell
// resolves, read before control moves. "If they do" is read after the
// change: the draw happens only when the opponent now controls the
// permanent, so a permanent or opponent that is no longer a legal target
// draws nothing (CR 608.2b).
//
// No simplification.
func init() {
	ninetyEighth := Mode("Follow the 98th Rule — Target opponent gains control of target nonland permanent you control. If they do, you draw three cards.",
		Clauses(TargetPlayer("target opponent", Opponent()),
			TargetPermanent("target nonland permanent you control", Nonland(), YouControl())))
	ninetyEighth.Purpose = game.Purpose{Draws: 3}
	Register(Spec{
		OracleID:     "5068eba6-ccf0-45f2-b0db-40f787ee3308",
		Name:         "Rules of Acquisition",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			Mode("Follow the 3rd Rule — Gain control of target nonland permanent an opponent controls. That opponent draws three cards.",
				TargetPermanent("target nonland permanent an opponent controls", Nonland(), OpponentControls())),
			ninetyEighth,
		),
		OnResolve: ResolveSteps(Instruction(rulesOfAcquisitionThird), Instruction(rulesOfAcquisitionNinetyEighth)),
	})
}

// rulesOfAcquisitionThird is the first bullet: take the permanent, then
// the opponent who controlled it draws three.
func rulesOfAcquisitionThird(ctx *Context) error {
	if !ctx.HasMode(0) {
		return nil
	}
	t, ok := ModeClauseTarget(ctx, 0, 0)
	if !ok || t.Kind != game.TargetCard {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(t.ID)
	if !ok {
		return nil
	}
	opponent := c.Controller
	if err := (GainControl{Target: t.ID, Duration: game.IndefiniteDuration(),
		Label: "Rules of Acquisition — gain control"}).Apply(ctx); err != nil {
		return err
	}
	return DrawCards{Player: opponent, N: 3}.Apply(ctx)
}

// rulesOfAcquisitionNinetyEighth is the second bullet: the opponent takes
// the permanent, and if they now control it, you draw three. Choose one
// means the chosen bullet is always occurrence 0 (TargetRef.Mode).
func rulesOfAcquisitionNinetyEighth(ctx *Context) error {
	if !ctx.HasMode(1) {
		return nil
	}
	opp, ok := ModeClauseTarget(ctx, 0, 0)
	if !ok || opp.Kind != game.TargetPlayer {
		return nil
	}
	perm, ok := ModeClauseTarget(ctx, 0, 1)
	if !ok || perm.Kind != game.TargetCard {
		return nil
	}
	if err := (GainControl{Target: perm.ID, Controller: opp.ID, Duration: game.IndefiniteDuration(),
		Label: "Rules of Acquisition — target opponent gains control"}).Apply(ctx); err != nil {
		return err
	}
	if c, ok := ctx.Game.LookupCardForEffect(perm.ID); !ok || c.Controller != opp.ID || !onBattlefield(ctx.Game, perm.ID) {
		return nil
	}
	return DrawCards{N: 3}.Apply(ctx)
}
