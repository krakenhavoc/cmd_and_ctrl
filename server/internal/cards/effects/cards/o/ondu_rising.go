package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ondu Rising — Sorcery {1}{W}:
//
//	"Whenever a creature attacks this turn, it gains lifelink until end of
//	 turn.
//	 Awaken 4—{4}{W}"
//
// ADR 0135 §3 (#2411): CR 603.7b's repeating delayed trigger keyed on
// EventAttack, once per attacking creature (CR 508.1), any player's, until
// the turn ends; each gives that creature lifelink until end of turn. Then
// the awaken land (CR 702.113a), which has haste and may attack this turn
// and gain lifelink too.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "01081f81-f588-43ab-8d47-13593aa19ce1",
		Name:         "Ondu Rising",
		Completeness: effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{
			effects.Awaken(4, "{4}{W}", nil),
		},
		OnResolve: effects.AwakenAfter(4, func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DelayedOnEvent{
				Label:   "Ondu Rising — it gains lifelink until end of turn",
				On:      []game.EventKind{game.EventAttack},
				Body:    onduRisingLifelinkBody,
				Repeats: true,
			}.Apply(ctx)
		}),
	})
}

// onduRisingLifelinkBody gives the creature that attacked (the trigger's
// payload) lifelink until end of turn. A creature that has left the
// battlefield by then gains nothing.
var onduRisingLifelinkBody = game.SimpleDelayedBody("ondu-rising/attacker-gains-lifelink", onduRisingLifelink)

func onduRisingLifelink(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, id := range ctx.PayloadCards() {
		if _, ok := g.PermanentRefForEffect(id); !ok {
			continue
		}
		if err := (effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"lifelink"}, Label: "Ondu Rising — lifelink"}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
