package cardsn

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Noxious Assault — Sorcery {3}{G}{G}:
//
//	"Creatures you control get +2/+2 until end of turn. Whenever a
//	 creature blocks this turn, its controller gets a poison counter."
//
// The pump is Makindi Stampede's, pinned to the creatures you control as
// the spell resolves (CR 611.2c). The second sentence is a delayed
// triggered ability (CR 603.7a, 603.7b) that watches blockers until the
// turn ends and triggers for each creature that blocks (CR 509.3a: once
// per blocker however many attackers it blocks), whoever controls it.
// Whose poison counter it is, is read off the blocker as it last existed
// (CR 608.2h), and the counter is given by the spell's controller
// (CR 120.3b's "you give") through the CR 614 player-counter window.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "95c0611a-b6f1-4cb8-8b47-73aa9c2a7d2c",
		Name:         "Noxious Assault",
		Completeness: effects.CompletenessFull,
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.BoostUntilEOT{
					Match:     effects.And(effects.Creature(), effects.YouControl()),
					Power:     2,
					Toughness: 2,
					Label:     "Noxious Assault — +2/+2",
				}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.DelayedOnEvent{
					Label:     "Noxious Assault — its controller gets a poison counter",
					On:        []game.EventKind{game.EventBlock},
					Condition: eachCreatureBlocksCondition,
					Body:      blockersControllerGetsAPoisonCounterBody,
					Repeats:   true,
				}.Apply(ctx)
			}),
		),
	})
}

var (
	// "Whenever a creature blocks this turn" (Noxious Assault): the
	// creature an EventBlock names, once however many attackers it blocks.
	eachCreatureBlocksCondition = game.DelayedEachCondition("block/each-creature-blocks", eachCreatureBlocks)

	// "Its controller gets a poison counter": the payload creature's
	// controller as it last existed.
	blockersControllerGetsAPoisonCounterBody = game.SimpleDelayedBody("poison/the-payload-creatures-controller", blockersControllerGetsAPoisonCounter)
)

func eachCreatureBlocks(ev game.Event, _ *game.DelayedTrigger, g *game.Game, _ game.EffectParams) []uuid.UUID {
	if ev.Kind != game.EventBlock || ev.CardID == uuid.Nil || ev.Target == uuid.Nil || ev.Amount > 1 {
		return nil
	}
	if c, ok := g.LookupCardForEffect(ev.CardID); !ok || !c.IsCreature() {
		return nil
	}
	return []uuid.UUID{ev.CardID}
}

func blockersControllerGetsAPoisonCounter(g *game.Game, item *game.StackItem) error {
	obj, ok := effects.PayloadObject(effects.NewContext(g, item))
	if !ok {
		return nil
	}
	info, ok := g.PermanentForEffect(obj)
	if !ok || info.Controller == uuid.Nil {
		return nil
	}
	return g.AddPlayerCounterByForEffect(item.Controller, info.Controller, game.CounterPoison, 1)
}
