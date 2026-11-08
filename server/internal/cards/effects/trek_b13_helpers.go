package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// trek_b13_helpers.go — Star Trek, batch 13 (#4973). Its own file, so the
// batch PRs landing in parallel do not meet in a shared one.

// CountersIfCastFrom is "If you do, it enters with a <kind> counter on
// it" after a card's own "You may cast this card from your <zone>"
// (Storied Veteran): a CR 614.1c clause read off the resolving spell's
// cast record (CastCounts.CastFrom), so it composes with every other
// entry clause. A cast from anywhere else, a copy of the spell (which was
// not cast, CR 707.10) and a permanent put onto the battlefield without
// being cast enter with none.
func CountersIfCastFrom(zone game.ZoneKind, kind string, n int) game.EntryCountersFromCast {
	return game.EntryCountersFromCast{
		Kind: kind,
		Count: func(cast game.CastCounts) int {
			if cast.CastFrom != zone {
				return 0
			}
			return n
		},
	}
}

// ReturnTheModesTargetsToHand is a modal bullet's "Return one or two
// target ... cards from your graveyard to your hand" (Yuna's Decision,
// Stray from History): each of this occurrence's targets that is still
// legal (CR 608.2b) returns, and one that has left stays where it is.
func ReturnTheModesTargetsToHand(_ *game.StackItem, ctx *Context, occ int) error {
	for _, t := range ctx.ModeTargets(occ) {
		if !ctx.IsTargetLegal(t) {
			continue
		}
		if err := (ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneHand}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// ReturnTheModesTargetToBattlefield is a modal bullet's "Return target
// ... card from your graveyard to the battlefield" (Inscription of Ruin,
// Resistance Is Futile): the occurrence's still-legal target returns under
// its owner's control.
func ReturnTheModesTargetToBattlefield(_ *game.StackItem, ctx *Context, occ int) error {
	t, ok := ModeTarget(ctx, occ)
	if !ok {
		return nil
	}
	return ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield}.Apply(ctx)
}

// ReturnTheModesTargetToBattlefieldTapped is a modal bullet's "Return
// target ... card from your graveyard to the battlefield tapped" (Phoenix
// Down, Deadly Plot): the occurrence's still-legal target returns tapped.
func ReturnTheModesTargetToBattlefieldTapped(_ *game.StackItem, ctx *Context, occ int) error {
	t, ok := ModeTarget(ctx, occ)
	if !ok {
		return nil
	}
	return ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield, Tapped: true}.Apply(ctx)
}
