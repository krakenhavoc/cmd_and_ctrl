package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jailbreak Scheme — Sorcery {U} (Outlaws of Thunder Junction):
//
//	"Spree (Choose one or more additional costs.)
//	 + {3} — Put a +1/+1 counter on target creature. It can't be blocked this
//	   turn.
//	 + {2} — Target artifact or creature's owner puts it on their choice of the
//	   top or bottom of their library."
//
// The first bullet is the counter, then Artful Dodge's restriction on the same
// creature; the second is Endless Detour's tuck: the OWNER picks the lane, and
// the move is the library route from the battlefield, so a commander's owner
// is offered the command zone (CR 903.9) and a token simply ceases to exist.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bc7f39e1-97bf-4725-a84a-2a04121530c6",
		Name:         "Jailbreak Scheme",
		Completeness: CompletenessFull,
		Modes: Spree(
			SpreeModeDoing("Put a +1/+1 counter on target creature. It can't be blocked this turn.", "{3}",
				TargetCreature("target creature"),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					if err := (AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
						return err
					}
					return RestrictUntilEOT{
						Target:       t.ID,
						Restrictions: game.CantBeBlocked,
						Label:        "Jailbreak Scheme — can't be blocked",
					}.Apply(ctx)
				}),
			SpreeModeDoing("Target artifact or creature's owner puts it on their choice of the top or bottom of their library.", "{2}",
				TargetPermanent("target artifact or creature", Or(Artifact(), Creature())),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					c, ok := ctx.Game.LookupCardForEffect(t.ID)
					if !ok || c.Owner == uuid.Nil {
						return nil
					}
					return PutInLibraryInAnyOrder{
						Chooser:   c.Owner,
						Cards:     []uuid.UUID{t.ID},
						Placement: game.LibraryPlaceTopOrBottom,
						Label:     "Jailbreak Scheme — put it on the top or bottom of your library",
					}.Apply(ctx)
				}),
		),
	})
}
