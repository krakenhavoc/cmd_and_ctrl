package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chronomantic Escape — Sorcery {4}{W}{W}:
//
//	"Until your next turn, creatures can't attack you. Exile
//	 Chronomantic Escape with three time counters on it.
//	 Suspend 3—{2}{W}"
//
// The first sentence is a CR 508.1c restriction on each opponent's
// attacks, not on the creatures now on the battlefield: a
// CantAttackGrant on every opponent (GrantCantAttackPlayerUntilYour-
// NextTurnForEffect) that covers the player only, so a planeswalker
// they control can still be attacked (the 2014-02-01 ruling), and a
// creature that enters later is refused too (the 2013-04-15 ruling).
// It lasts until your next turn begins (CR 611.2).
//
// The second sentence is Rousing Refrain's: the spell exiles itself
// and the three time counters go on the exiled card, which suspend's
// upkeep trigger then counts down (CR 702.62a), so the card recurs.
//
// No simplification. Star Trek Commander (#4974).
func init() {
	const name = "Chronomantic Escape"
	Register(Spec{
		OracleID:       "629133e2-af20-4107-bc82-11ba32d5e2a6",
		Name:           name,
		Completeness:   CompletenessFull,
		SpecialActions: []game.SpecialAction{Suspend(3, "{2}{W}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, opp := range ctx.Opponents() {
				ctx.Game.GrantCantAttackPlayerUntilYourNextTurnForEffect(opp, item.Controller,
					game.CantAttackScope{PlayerOnly: true},
					name+" — until your next turn, creatures can't attack you", item.SourceCardID)
			}
			return ExileTarget{Target: item.SourceCardID, ThenTail: ctx.ItemTail(rousingRefrainSuspended)}.Apply(ctx)
		},
	})
}
