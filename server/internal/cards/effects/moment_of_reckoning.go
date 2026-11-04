package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Moment of Reckoning — Sorcery {3}{W}{W}{B}{B} (Secrets of Strixhaven,
// #4456):
//
//	"Choose up to four. You may choose the same mode more than once.
//	 • Destroy target nonland permanent.
//	 • Return target nonland permanent card from your graveyard to the
//	   battlefield."
//
// Zero to four choices, repeats allowed (CR 700.2d, ChooseNRepeating with a
// floor of zero), each choice with its own target; a mode chosen several
// times is treated as appearing that many times in sequence, and the modes
// are followed in the order written (CR 608.2c). A return aimed at a card
// that is gone from the graveyard by then, because an earlier choice of the
// same mode already returned it, does nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8e9576c1-b0c2-4be5-9005-35b51164f78e",
		Name:         "Moment of Reckoning",
		Completeness: CompletenessFull,
		Modes: ChooseNRepeating("Choose up to four (you may choose the same mode more than once)", 0, 4,
			ModeDoing("Destroy target nonland permanent.",
				TargetPermanent("target nonland permanent", Nonland()),
				DestroyTheModesTarget),
			ModeDoing("Return target nonland permanent card from your graveyard to the battlefield.",
				TargetCardInGraveyard("target nonland permanent card in your graveyard", YouOwn(), Permanent(), Nonland()),
				momentOfReckoningReturn),
		),
	})
}

// momentOfReckoningReturn returns the chosen card to the battlefield under
// your control.
func momentOfReckoningReturn(_ *game.StackItem, ctx *Context, occ int) error {
	t, ok := ModeTarget(ctx, occ)
	if !ok {
		return nil
	}
	return ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield, Controller: ctx.Controller()}.Apply(ctx)
}
