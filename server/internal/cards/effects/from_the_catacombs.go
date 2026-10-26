package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// From the Catacombs — Sorcery {3}{B}{B}:
//
//	"Put target creature card from a graveyard onto the battlefield
//	 under your control with a corpse counter on it. You take the
//	 initiative. If that creature would leave the battlefield, exile it
//	 instead of putting it anywhere else.
//	 Escape—{3}{B}{B}, Exile five other cards from your graveyard."
//
// Requested for the Captain N'ghathrod deck (#4092); waited on the
// initiative (#4096, ADR 0203).
//
//   - "From a graveyard … under your control": any graveyard, and the
//     creature changes hands — ReturnFromGraveyardWithCountersThen with
//     the caster as controller, as Reanimate does.
//   - "With a corpse counter on it": the counter rides the entry (CR
//     614.1c), so it is on the creature as it enters, through the same
//     counter replacements any other entering counter meets.
//   - "You take the initiative": TakeTheInitiative. The venture into
//     Undercity it causes is the initiative's own triggered ability, so
//     it goes on the stack after the sorcery has finished resolving.
//   - "If that creature would leave the battlefield, exile it instead":
//     the per-object replacement Whip of Erebos and unearth share
//     (ExileInsteadOfLeavingBattlefield), registered once the creature
//     has entered — a creature whose entry first asked a question (a
//     Clone's copy) still gets it, because it is registered from the
//     entry's continuation.
//   - Escape: Escape("{3}{B}{B}", 5) from the graveyard (CastableZones).
//     An escaped sorcery goes to the graveyard as it resolves, like any
//     other, and may escape again — escape exiles nothing but its cost.
//
// The sentences resolve in printed order. If the target is gone, the
// spell does not resolve at all (CR 608.2b), so nobody takes the
// initiative.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "cc21b29f-2d1a-4769-9b65-f92bd2ca1cc8",
		Name:             "From the Catacombs",
		Completeness:     CompletenessFull,
		Targets:          TargetCardInGraveyard("target creature card from a graveyard", Creature()),
		Purpose:          ForTargets(Reanimates(0)),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Escape("{3}{B}{B}", 5)},
		OnResolve:        fromTheCatacombsResolve,
	})
}

// fromTheCatacombsResolve is the printed text, in order.
func fromTheCatacombsResolve(item *game.StackItem, ctx *Context) error {
	you := ctx.Controller()
	targets := ctx.LegalTargets()
	if len(targets) == 0 || targets[0].Kind != game.TargetCard {
		return nil
	}
	if err := ctx.Game.ReturnFromGraveyardWithCountersKeyedForEffect(targets[0].ID, you, false,
		map[string]int{"corpse": 1}, ctx.Tail(fromTheCatacombsExileInstead)); err != nil {
		return nothingIfGone(err)
	}
	return TakeTheInitiative{Player: you}.Apply(ctx)
}

// fromTheCatacombsExileInstead is the entry's continuation, a keyed tail
// (#4253 slice 6): the creature that entered gets the leave-the-battlefield
// redirect, for the tail's Player. An on-disk key.
var fromTheCatacombsExileInstead = EnteredTail("from-the-catacombs/exile-instead", func(ctx *Context, entered uuid.UUID, o game.TailOutcome) error {
	if entered == uuid.Nil {
		return nil
	}
	ExileInsteadOfLeavingBattlefield(ctx.Game, entered, o.Tail.Player,
		"From the Catacombs: if that creature would leave the battlefield, exile it instead")
	return nil
})
