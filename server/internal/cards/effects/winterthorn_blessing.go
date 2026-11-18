package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Winterthorn Blessing — Sorcery {G}{U} (Midnight Hunt, #5497):
//
//	"Put a +1/+1 counter on up to one target creature you control. Tap up
//	 to one target creature you don't control, and that creature doesn't
//	 untap during its controller's next untap step.
//	 Flashback {1}{G}{U} (You may cast this card from your graveyard for
//	 its flashback cost. Then exile it.)"
//
// Two independent "up to one" clauses: either can be empty, and each half
// resolves on its own if the other target became illegal (CR 608.2b).
// TapAndFreeze is the shared tap-and-hold resolution; the hold is the
// target's own controller's next untap step.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4b4bdbdc-3934-471b-9027-31277e7765ab",
		Name:         "Winterthorn Blessing",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetCreature("up to one target creature you control", YouControl()).WithCount(0, 1),
			TargetCreature("up to one target creature you don't control", Not(YouControl())).WithCount(0, 1),
		),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{1}{G}{U}")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetCard && ctx.IsTargetLegal(t) {
				if err := (AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
					return err
				}
			}
			if t, ok := ctx.ClauseTarget(1); ok && t.Kind == game.TargetCard && ctx.IsTargetLegal(t) {
				return TapAndFreeze{
					Targets: []uuid.UUID{t.ID},
					Label:   "Winterthorn Blessing — doesn't untap during its controller's next untap step",
				}.Apply(ctx)
			}
			return nil
		},
	})
}
