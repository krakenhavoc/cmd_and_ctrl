package effects

import (
	"sort"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// delayed_that_creature_dies.go — "When that creature dies this turn,
// <do something with a target>" (#2057):
//
//	"When that creature dies this turn, put its counters on up to one
//	 target creature you control and draw a card."   Heroic Sacrifice
//	"When that creature dies this turn, you earthbend 4."  Fatal Fissure
//
// A delayed triggered ability the resolving spell creates (CR 603.7a),
// about one creature as the object it is now. It is an each-condition
// (game.DelayedEachCondition) keyed by CondParams.Object, the creature's
// instance ID and CR 400.7 epoch, so a creature that left and came back
// is a new object and never fires it (CR 603.7c), and the fired
// instance's payload is that object with its epoch: "its counters" are
// read from the record of it as it last existed on the battlefield (CR
// 608.2h, Game.PermanentForEffect). A one-shot trigger, so it triggers
// once (CR 603.7b); the default "this turn" duration ends it at cleanup
// whether or not it fired (CR 514.2).
//
// The body is registered with its target clause (game.TargetedDelayedBody,
// ADR 0187 amendment 2026-10-16), chosen as the instance goes on the
// stack (CR 603.3d). Whippoorwill and Together Forever keep the older
// untargeted shape, which lists the creature on Cards.

var (
	// "When that creature dies this turn": the one creature named by
	// CondParams.Object, put into a graveyard from the battlefield while
	// it was that object and a creature (CR 700.4, 603.10a).
	thatCreatureDiedCondition = game.DelayedEachCondition("dies/that-creature", thatCreatureDied)

	// "Put its counters on up to one target creature you control and draw
	// a card" (Heroic Sacrifice, #2057).
	heroicSacrificeDiesBody = game.TargetedDelayedBody("heroic-sacrifice/its-counters-and-draw",
		simpleBody(heroicSacrificeDies),
		constTargets(func() *game.TargetSpec {
			return TargetCreature("up to one target creature you control", YouControl()).WithCount(0, 1)
		}))

	// "You earthbend 4" (Fatal Fissure, #2057): earthbend's own "target
	// land you control".
	fatalFissureDiesBody = game.TargetedDelayedBody("fatal-fissure/earthbend-4",
		simpleBody(fatalFissureDies), constTargets(EarthbendTargets))
)

// thatCreatureDied matches the EventLTB of the creature CondParams.Object
// names, to a graveyard, when the object that left is the one the
// trigger was created about. The epoch of the object that left is its
// last-known record's, written as it left.
func thatCreatureDied(ev game.Event, _ *game.DelayedTrigger, g *game.Game, p game.EffectParams) []uuid.UUID {
	if p.Object.ID == uuid.Nil || ev.CardID != p.Object.ID {
		return nil
	}
	if _, ok := diedCreature(ev, g); !ok {
		return nil
	}
	ref, ok := g.PermanentRefForEffect(ev.CardID)
	if !ok || ref.Epoch != p.Object.Epoch {
		return nil
	}
	return []uuid.UUID{ev.CardID}
}

// whenThatCreatureDiesThisTurn creates "when <creature> dies this turn,
// <body>" about the creature as the object it is now, controlled by the
// resolving object's controller (CR 603.7d). A creature that is no
// longer on the battlefield creates nothing: the trigger could never
// fire.
func whenThatCreatureDiesThisTurn(ctx *Context, creature uuid.UUID, label string, body game.BodyRef) error {
	if creature == uuid.Nil {
		return nil
	}
	c, ok := battlefieldCardForEffect(ctx.Game, creature)
	if !ok {
		return nil
	}
	return DelayedOnEvent{
		Label:      label,
		On:         []game.EventKind{game.EventLTB},
		Condition:  thatCreatureDiedCondition,
		CondParams: game.EffectParams{Object: game.ObjectRef{ID: creature, Epoch: c.ObjectEpoch}},
		Body:       body,
	}.Apply(ctx)
}

// heroicSacrificeDies is the fired trigger's body: every counter the
// creature had as it last existed on the battlefield (CR 608.2h), kind by
// kind, on the chosen creature, then a card. With no target chosen, or a
// creature that had no counters, it only draws. A chosen target that is
// illegal as it resolves is the only target, so the ability does not
// resolve at all (CR 608.2b) and draws nothing.
//
// Each kind is a new placement through the CR 614 counter window
// (Doubling Season, Hardened Scales), put by the trigger's controller
// (CR 120.3d), and the draw waits for the last of them to land, so a
// CR 616 ordering prompt does not let the draw go first.
func heroicSacrificeDies(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	controller := ctx.Controller()
	draw := func(g *game.Game) error {
		return DrawCards{Player: controller, N: 1}.Apply(NewContext(g, item))
	}
	target := firstLegalCardTarget(ctx)
	obj, ok := payloadObject(ctx)
	if target == uuid.Nil || !ok || !onBattlefield(g, target) {
		return draw(g)
	}
	info, ok := g.PermanentForEffect(obj)
	if !ok || len(info.Counters) == 0 {
		return draw(g)
	}
	return putCountersThen(g, controller, target, info.Counters, draw)
}

// putCountersThen puts each kind in `counters` on `target`, in a fixed
// order so a log reads the same twice, one placement after another, and
// runs `then` once the last has landed.
func putCountersThen(g *game.Game, placer, target uuid.UUID, counters map[string]int, then func(g *game.Game) error) error {
	kinds := make([]string, 0, len(counters))
	for k, n := range counters {
		if n > 0 {
			kinds = append(kinds, k)
		}
	}
	sort.Strings(kinds)
	var next func(g *game.Game, i int) error
	next = func(g *game.Game, i int) error {
		if i == len(kinds) {
			return then(g)
		}
		return g.AddCounterByThenForEffect(placer, target, kinds[i], counters[kinds[i]], func(g *game.Game, _ int) error {
			return next(g, i+1)
		})
	}
	return next(g, 0)
}

// fatalFissureDies is the fired trigger's body: earthbend 4 on the land
// it targeted.
func fatalFissureDies(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	land := FirstLegalBattlefieldTarget(ctx)
	if land == uuid.Nil {
		return nil
	}
	return Earthbend{Target: land, N: 4}.Apply(ctx)
}
