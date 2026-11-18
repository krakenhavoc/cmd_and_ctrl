package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kros, Defense Contractor — Legendary Creature — Cat Advisor {1}{G}{W}{U},
// 2/4 (Streets of New Capenna Commander):
//
//	"At the beginning of your upkeep, put a shield counter on target
//	 creature an opponent controls.
//	 Whenever you put one or more counters on a creature you don't
//	 control, tap that creature and goad it. It gains trample until your
//	 next turn."
//
// The first ability targets. The second is one trigger per creature per
// batch of counters that Kros's controller put
// (WheneverYouPutOneOrMoreCountersOnACreature), so the upkeep's own shield
// counter triggers it. "That creature" is stamped as
// the object it was when the counters landed (CR 400.7): one that left
// and returned meanwhile is not tapped, goaded or given trample. The goad
// and the trample both last until Kros's controller's next turn.
//
// No simplification.
func init() {
	const label = "Kros, Defense Contractor — tap that creature and goad it; it gains trample until your next turn"
	counters := WheneverYouPutOneOrMoreCountersOnACreature(creatureYouDontControl, label, krosTapGoad)
	counters.Build = func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, label)
		item.Params.Object = game.ObjectRef{ID: ev.Target, Epoch: g.ObjectEpochForEffect(ev.Target)}
		return item
	}
	Register(Spec{
		OracleID:     "debeb004-3157-43cf-95fd-220777417ff7",
		Name:         "Kros, Defense Contractor",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				AtYourUpkeep("Kros, Defense Contractor — put a shield counter on target creature an opponent controls",
					Do(Instruction(func(ctx *Context) error {
						t, ok := ctx.ClauseTarget(0)
						if !ok {
							return nil
						}
						return AddCounter{Target: t.ID, Kind: game.CounterShield, N: 1}.Apply(ctx)
					}))),
				TargetCreature("target creature an opponent controls", OpponentControls())),
			counters,
		},
	})
}

// krosTapGoad is the second ability's body: tap, goad and trample the
// stamped creature while it is still that object on the battlefield.
func krosTapGoad(g *game.Game, item *game.StackItem) error {
	obj := item.Params.Object
	if obj.ID == uuid.Nil || !onBattlefield(g, obj.ID) || g.ObjectEpochForEffect(obj.ID) != obj.Epoch {
		return nil
	}
	ctx := NewContext(g, item)
	if err := g.TapTargetForEffect(obj.ID); err != nil {
		return err
	}
	if err := goadAndScheduleClear(ctx, []uuid.UUID{obj.ID}); err != nil {
		return err
	}
	return ScopedEffectFor{
		Target:   obj.ID,
		Mods:     []game.Mod{game.AddKeywordsMod("trample")},
		Duration: DurationUntilYourNextTurn(ctx, ctx.Controller()),
		Label:    "Kros, Defense Contractor — trample until your next turn",
	}.Apply(ctx)
}
