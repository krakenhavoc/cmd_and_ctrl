package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Experimental Lab // Staff Room — Enchantment — Room (ADR 0103):
//
//	Experimental Lab {3}{G}: "When you unlock this door, manifest dread,
//	 then put two +1/+1 counters and a trample counter on that creature."
//	Staff Room {2}{G}: "Whenever a creature you control deals combat
//	 damage to a player, turn that creature face up or put a +1/+1
//	 counter on it."
//
// Experimental Lab is manifest dread (CR 701.62a, ADR 0082's
// 2026-10-07 amendment) with its counters as the continuation: two
// +1/+1 counters and a trample counter on the creature that entered.
//
// Staff Room is a choice, asked only when there is one. A face-up
// creature (the usual case) can only take the counter, so it gets it
// without a prompt. A face-down creature that the rules let an effect
// turn over (game.CanTurnFaceUpForEffect, ADR 0082's second 2026-10-07
// amendment, #2590) is offered both: turn it face up, which costs
// nothing, or the counter. A manifested noncreature card cannot be
// turned face up (CR 701.40b), so it is not offered the turn and takes
// the counter.
func init() {
	Register(Room(RoomSpec{
		OracleID:     "8869df0c-fe84-4964-8d38-a8ecefa9c252",
		Name:         "Experimental Lab // Staff Room",
		Completeness: CompletenessFull,
		Left: Door{Triggered: []game.TriggeredAbility{
			WhenYouUnlockThisDoor(game.DoorLeft, "Experimental Lab — manifest dread, then put two +1/+1 counters and a trample counter on that creature",
				Do(ManifestDread{Then: PutCountersOnManifested(
					CounterAmount{Kind: game.CounterPlusOne, N: 2},
					CounterAmount{Kind: game.CounterTrample, N: 1},
				)})),
		}},
		Right: Door{Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return combatDamageToPlayerBy(ev, source.Controller, g)
			}, "Staff Room — turn that creature face up or put a +1/+1 counter on it",
				staffRoomTurnUpOrCounter),
		}},
	}))
}

// staffRoomTurnUpOrCounter is "turn that creature face up or put a
// +1/+1 counter on it" for the creature that dealt the damage.
func staffRoomTurnUpOrCounter(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	id := item.Trigger.Event.Source
	if !b15OnBattlefield(g, id) {
		return nil
	}
	ctx := NewContext(g, item)
	counter := func(ctx *Context) error {
		return AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
	}
	if !canTurnFaceUpByEffect(g, id) {
		return counter(ctx)
	}
	return MayChoice{
		Question: "Staff Room — turn that creature face up, or put a +1/+1 counter on it?",
		YesLabel: "Turn it face up",
		NoLabel:  "Put a +1/+1 counter",
		OnYes: func(ctx *Context) error {
			return TurnFaceUp{Targets: []uuid.UUID{id}}.Apply(ctx)
		},
		OnNo: counter,
	}.Apply(ctx)
}
