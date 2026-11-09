package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Desperate Measures — Instant {B} (Tarkir: Dragonstorm, #4463):
//
//	"Target creature gets +1/-1 until end of turn. When it dies under
//	 your control this turn, draw two cards."
//
// The pump is a layer-7c boost until end of turn, then the delayed
// triggered ability of A Good Day to Die's family (whenThatCreatureDies
// ThisTurnIf), about that one object (CR 603.7c). The condition adds the
// controller test: the creature must have died under the spell's
// controller's control, read from the exit's last-known information
// (CR 603.10a). A creature an opponent controls when it dies does not
// draw; one that was stolen from them and then dies under yours does.
// A creature that is already gone as the spell resolves creates nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f1f3df1c-5c51-445a-b66d-eee4aab23691",
		Name:         "Desperate Measures",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || !onBattlefield(ctx.Game, t.ID) {
				return nil
			}
			if err := (BoostUntilEOT{Target: t.ID, Power: 1, Toughness: -1, Label: "Desperate Measures — +1/-1"}).Apply(ctx); err != nil {
				return err
			}
			return whenThatCreatureDiesThisTurnIf(ctx, t.ID, "Desperate Measures — draw two cards",
				desperateMeasuresDiedCondition, desperateMeasuresDrawBody)
		},
	})
}

// desperateMeasuresDiedCondition is thatCreatureDied narrowed to "under
// your control": the creature's last-known controller is the delayed
// trigger's controller. An on-disk key.
var desperateMeasuresDiedCondition = game.DelayedEachCondition("dies/that-creature-under-your-control",
	func(ev game.Event, dt *game.DelayedTrigger, g *game.Game, p game.EffectParams) []uuid.UUID {
		ids := thatCreatureDied(ev, dt, g, p)
		if len(ids) == 0 {
			return nil
		}
		c, ok := g.LookupCardForEffect(ev.CardID)
		if !ok || leftUnderControlOf(ev, c) != dt.Controller {
			return nil
		}
		return ids
	})

// desperateMeasuresDrawBody is "draw two cards". An on-disk key.
var desperateMeasuresDrawBody = game.SimpleDelayedBody("desperate-measures/draw-two", func(g *game.Game, item *game.StackItem) error {
	return DrawCards{Player: item.Controller, N: 2}.Apply(NewContext(g, item))
})
