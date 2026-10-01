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
// DECLARED SIMPLIFICATION on Pull, weaker than printed: the two cards
// enter one after the other, as every multi-card return from a
// graveyard in the catalog does (Reveillark, Lich-Knights' Conquest),
// rather than together. The engine's simultaneous entry door takes
// cards from a hand, a library, exile or the command zone, not from a
// graveyard. The difference shows only with an ability that watches
// other creatures enter: the first creature sees the second enter, and
// the second does not see the first.
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
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The two creatures enter one after the other rather than at the same time, so an ability that watches creatures enter may see only one of them."},
		Targets:      UpToCardsFromASingleGraveyard("up to two target creature cards from a single graveyard", 2, Creature()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			var entered []uuid.UUID
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield, Controller: ctx.Controller()}).Apply(ctx); err != nil {
					return err
				}
				if onBattlefield(ctx.Game, t.ID) {
					entered = append(entered, t.ID)
				}
			}
			if len(entered) == 0 {
				return nil
			}
			for _, id := range entered {
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
				Cards: entered,
				Body:  sacrificeListedCardsBody,
			}.Apply(ctx)
		},
	})
}
