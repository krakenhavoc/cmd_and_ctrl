package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Push // Pull — split card, oracle 7e050495-bed7-43b9-abce-866c61beb1da:
//
//	Push — Sorcery {1}{W/B}: "Destroy target tapped creature."
//	Pull — Sorcery {4}{B/R}{B/R}: "Put up to two target creature cards
//	       from a single graveyard onto the battlefield under your
//	       control. They gain haste until end of turn. Sacrifice them
//	       at the beginning of the next end step."
//
// #1807, ADR 0106 §5. Each half is an ADR 0034 face (ADR 0103): Push
// under the bare oracle ID, Pull under "#1". Only one half can be cast
// (CR 709.3).
//
// Pull's creatures enter under the caster's control whoever owns them,
// get haste, and are sacrificed by a delayed trigger at the next end
// step; one that has left, or that the caster no longer controls by
// then, is not sacrificed.
//
// The two cards enter together, as one event (#1867): each sees the
// other enter (CR 603.6a), so a "whenever another creature enters"
// ability on either one triggers for the other.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7e050495-bed7-43b9-abce-866c61beb1da",
		Name:         "Push",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target tapped creature", tappedPermanent()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return destroyChosenPermanent(ctx.Game, item)
		},
	})
	Register(Spec{
		OracleID:     "7e050495-bed7-43b9-abce-866c61beb1da#1",
		Name:         "Pull",
		Completeness: CompletenessFull,
		Targets:      UpToCardsFromASingleGraveyard("up to two target creature cards from a single graveyard", 2, Creature()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return ReturnFromGraveyardTogether{
				Targets:    legalTargetCardIDs(ctx),
				Controller: ctx.Controller(),
				Then:       pullHasteThenSacrifice,
			}.Apply(ctx)
		},
	})
}

// pullHasteThenSacrifice is the rest of Pull, once its creatures have
// entered: haste until end of turn, and a delayed trigger that
// sacrifices them at the beginning of the next end step.
func pullHasteThenSacrifice(ctx *Context, entered []uuid.UUID) error {
	var here []uuid.UUID
	for _, id := range entered {
		if onBattlefield(ctx.Game, id) {
			here = append(here, id)
		}
	}
	if len(here) == 0 {
		return nil
	}
	for _, id := range here {
		if err := (GrantKeywordUntilEOT{
			Target:   id,
			Keywords: []string{"haste"},
			Label:    "Pull — haste until end of turn",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return ScheduleDelayedTrigger{
		Label: "Pull — sacrifice them",
		Cards: here,
		Body:  sacrificeListedCardsBody,
	}.Apply(ctx)
}
