package game

import (
	"maps"

	"github.com/google/uuid"
)

// life_batch.go — how much life each player lost in ONE event batch
// (#2183, CR 603.2c, CR 120.3a).
//
// "Whenever one or more opponents each lose exactly 1 life" (Ob
// Nixilis, Captive Kingpin) reads a TOTAL: the life one player lost to
// everything that happened at the same time. The engine emits one
// event per source — one EventDealDamage per creature, one
// EventChangeLife per loss — so two creatures dealing 1 combat damage
// each to one opponent are two events of 1 and one loss of 2. Reading
// the events one at a time would trigger on both, stronger than
// printed. OncePerBatch (event_batch.go) collapses WHICH events make a
// trigger, and says nothing about how much they add up to.
//
// # The two halves
//
//   - batchLifeLost is the running total per player for the live
//     batch, written at the three places a life total actually moves
//     (both arms of applyResolvedDamageToPlayerLocked and
//     applyResolvedLifeChangeLocked), so it counts what was LOST —
//     after every CR 614 replacement (a damage doubler changes it),
//     and not what an infect or "life total can't change" player
//     never lost. Life paid as a cost goes through the same pipeline
//     and is a loss (CR 119.4), so it counts.
//   - A TriggeredAbility with AtBatchEnd set is NOT dispatched when its
//     event arrives: the total is still growing. The harvester stages
//     it (stagedBatchTrigger) and settleBatchEndTriggersLocked asks
//     AtBatchEnd once the batch is over, at the priority-grant
//     boundary (runStateChecksLocked) or as the next batch opens —
//     whichever comes first, and always with the batch still live.
//
// A batch is the unit event_batch.go defines: one resolution, one
// combat damage step (first strike and regular are two, CR 510.4), one
// turn-based action. Two separate resolutions are two batches, so two
// Blood Artist drains in a row are two losses of exactly 1, each its
// own occurrence.
//
// Staged triggers and the totals ride Clone / RestoreFrom (undo) but a
// persisted snapshot drops them: both exist only while a resolution is
// paused on one of its own prompts, and a server restart in that
// window loses the staged trigger — weaker than printed, never
// stronger.

// batchLifeLossTotals is the running per-player life lost in the batch
// named by Batch.
type batchLifeLossTotals struct {
	Batch uint64
	Lost  map[uuid.UUID]int
}

// stagedBatchTrigger is a trigger waiting for its batch to finish.
// Data only: the ability is found again by Key on the source, so
// nothing here is a closure and a clone shares the slice safely.
type stagedBatchTrigger struct {
	Event  Event
	Source uuid.UUID
	Key    string
}

// noteLifeLostLocked records n life lost by player in the live batch.
// n <= 0 is nothing lost. Caller must hold g.mu in write mode.
func (g *Game) noteLifeLostLocked(player uuid.UUID, n int) {
	if n <= 0 || player == uuid.Nil {
		return
	}
	b := g.currentEventBatchLocked()
	if g.batchLifeLost.Batch != b || g.batchLifeLost.Lost == nil {
		g.batchLifeLost = batchLifeLossTotals{Batch: b, Lost: map[uuid.UUID]int{}}
	}
	g.batchLifeLost.Lost[player] += n
}

// lifeLostInBatchLocked is the total life player has lost in the live
// event batch. Caller must hold g.mu.
func (g *Game) lifeLostInBatchLocked(player uuid.UUID) int {
	if g.batchLifeLost.Batch != g.currentEventBatchLocked() {
		return 0
	}
	return g.batchLifeLost.Lost[player]
}

// LifeLostInBatchForEffect is the total life `player` has lost in the
// live event batch — everything lost at the same time as the event a
// trigger is being asked about (CR 603.2c). Meant for a
// TriggeredAbility.AtBatchEnd condition. Caller must hold g.mu (the
// trigger conditions already do).
func (g *Game) LifeLostInBatchForEffect(player uuid.UUID) int {
	return g.lifeLostInBatchLocked(player)
}

func cloneBatchLifeLossTotals(in batchLifeLossTotals) batchLifeLossTotals {
	if len(in.Lost) == 0 {
		return batchLifeLossTotals{Batch: in.Batch}
	}
	return batchLifeLossTotals{Batch: in.Batch, Lost: maps.Clone(in.Lost)}
}

// stageBatchEndTriggerLocked parks a matched AtBatchEnd ability until
// the batch settles. One entry per (source, key, batch): a second
// matching event of the same batch adds nothing, because the question
// is asked once, of the final totals. Caller must hold g.mu.
func (g *Game) stageBatchEndTriggerLocked(ev Event, source uuid.UUID, key string) {
	for _, s := range g.stagedBatchTriggers {
		if s.Source == source && s.Key == key && s.Event.Batch == ev.Batch {
			return
		}
	}
	g.stagedBatchTriggers = append(g.stagedBatchTriggers[:len(g.stagedBatchTriggers):len(g.stagedBatchTriggers)],
		stagedBatchTrigger{Event: ev, Source: source, Key: key})
}

// settleBatchEndTriggersLocked asks every staged ability whether its
// batch's totals satisfy it and dispatches the ones that do, through
// the ordinary harvest path (suppressors, OncePerBatch and trigger
// doublers all still apply). Idempotent and cheap when nothing is
// staged. The source must still be on the battlefield: the settle runs
// before the state-based actions sweep, so a creature a combat damage
// step is about to kill is still there to trigger.
//
// Caller must hold g.mu in write mode.
func (g *Game) settleBatchEndTriggersLocked() {
	if len(g.stagedBatchTriggers) == 0 {
		return
	}
	staged := g.stagedBatchTriggers
	g.stagedBatchTriggers = nil
	for _, s := range staged {
		if g.Battlefield == nil || !g.Battlefield.Contains(s.Source) {
			continue
		}
		source := g.findCardByIDLocked(s.Source)
		if source == nil {
			continue
		}
		src := *source
		lki := src.Effective()
		for _, t := range triggersOf(&src) {
			if t.Key != s.Key || t.AtBatchEnd == nil {
				continue
			}
			if !t.AtBatchEnd.holds(&src, g) {
				continue
			}
			pass := g.newHarvestPassLocked(s.Event)
			pass.settled = true
			g.harvestMatchLocked(&pass, src, lki, t, triggerOfPermanent)
		}
	}
}

// BatchEndCondition is what an AtBatchEnd ability asks of the settled
// batch. One reading so far, the one #2183 needs.
type BatchEndCondition struct {
	// OpponentLostExactly, when positive: at least one opponent of the
	// source's controller lost exactly this much life in the batch.
	// Eliminated seats do not count.
	OpponentLostExactly int
}

func (c *BatchEndCondition) holds(source *Card, g *Game) bool {
	if c.OpponentLostExactly > 0 {
		for _, seat := range g.Seats {
			if seat == nil || seat.Eliminated || seat.ID == source.Controller {
				continue
			}
			if g.lifeLostInBatchLocked(seat.ID) == c.OpponentLostExactly {
				return true
			}
		}
		return false
	}
	return false
}
