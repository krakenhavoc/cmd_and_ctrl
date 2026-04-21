package game

import "github.com/google/uuid"

// clone.go gives Game a deep-copy snapshot used by the undo stack
// and (later) replay machinery. The exported Snapshot already
// returns a shallow copy fine for read-only inspection; Clone goes
// further and gives back an entirely independent *Game whose
// mutation cannot affect the receiver.
//
// Cloned state intentionally drops the rwmutex (a fresh receiver
// owns its own) but preserves the captured rng pointer — so
// re-applying a chain of actions on the clone uses the same
// random source the original did, which is what tests expect.

// Clone returns a deep copy of the game suitable for stashing on the
// undo stack and later replacing the live receiver via *g = *clone.
//
// Concurrency: takes a read lock for the duration of the copy, so
// callers don't have to. The clone itself has its own zero-value
// mutex and is safe to mutate without affecting the original.
func (g *Game) Clone() *Game {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.cloneLocked()
}

// cloneLocked is the unlocked variant. Caller must already hold
// g.mu (read or write).
func (g *Game) cloneLocked() *Game {
	out := &Game{
		ID:                g.ID,
		CreatedAt:         g.CreatedAt,
		State:             g.State,
		Turn:              g.Turn,
		MulligansOpen:     g.MulligansOpen,
		Monarch:           g.Monarch,
		Initiative:        g.Initiative,
		UndoLimit:         g.UndoLimit,
		StartingSeat:      g.StartingSeat,
		SplitSecondActive: g.SplitSecondActive,
		rng:               g.rng,
	}
	if len(g.StackMeta) > 0 {
		out.StackMeta = make(map[uuid.UUID]*StackItem, len(g.StackMeta))
		for k, v := range g.StackMeta {
			out.StackMeta[k] = cloneStackItem(v)
		}
	}
	if len(g.PendingTriggers) > 0 {
		out.PendingTriggers = make([]*StackItem, len(g.PendingTriggers))
		for i, t := range g.PendingTriggers {
			out.PendingTriggers[i] = cloneStackItem(t)
		}
	}
	if len(g.LoyaltyActivatedThisTurn) > 0 {
		out.LoyaltyActivatedThisTurn = make(map[uuid.UUID]bool, len(g.LoyaltyActivatedThisTurn))
		for k, v := range g.LoyaltyActivatedThisTurn {
			out.LoyaltyActivatedThisTurn[k] = v
		}
	}
	if len(g.DiscardPending) > 0 {
		out.DiscardPending = make(map[uuid.UUID]int, len(g.DiscardPending))
		for k, v := range g.DiscardPending {
			out.DiscardPending[k] = v
		}
	}
	out.Battlefield = cloneZone(g.Battlefield)
	out.Stack = cloneZone(g.Stack)
	out.Exile = cloneZone(g.Exile)
	out.Seats = make([]*Player, len(g.Seats))
	for i, p := range g.Seats {
		out.Seats[i] = clonePlayer(p)
	}
	if len(g.Promises) > 0 {
		out.Promises = make(map[PromiseKey]int, len(g.Promises))
		for k, v := range g.Promises {
			out.Promises[k] = v
		}
	}
	if g.Vote != nil {
		out.Vote = cloneVote(g.Vote)
	}
	return out
}

func cloneZone(z *Zone) *Zone {
	if z == nil {
		return nil
	}
	cloned := &Zone{Kind: z.Kind, Owner: z.Owner}
	if len(z.Cards) > 0 {
		cloned.Cards = make([]Card, len(z.Cards))
		for i, c := range z.Cards {
			cloned.Cards[i] = cloneCard(c)
		}
	}
	return cloned
}

func cloneCard(c Card) Card {
	out := c
	if len(c.Counters) > 0 {
		out.Counters = make(map[string]int, len(c.Counters))
		for k, v := range c.Counters {
			out.Counters[k] = v
		}
	} else {
		out.Counters = nil
	}
	return out
}

func clonePlayer(p *Player) *Player {
	out := &Player{
		ID:                p.ID,
		Name:              p.Name,
		Seat:              p.Seat,
		Life:              p.Life,
		Poison:            p.Poison,
		Energy:            p.Energy,
		Eliminated:        p.Eliminated,
		HandKept:          p.HandKept,
		MulligansTaken:    p.MulligansTaken,
		DeckImported:      p.DeckImported,
		UndosRemaining:    p.UndosRemaining,
		DiscordID:         p.DiscordID,
		DiscordAvatarHash: p.DiscordAvatarHash,
		DisplayName:       p.DisplayName,
	}
	out.Library = cloneZone(p.Library)
	out.Hand = cloneZone(p.Hand)
	out.Graveyard = cloneZone(p.Graveyard)
	out.Command = cloneZone(p.Command)
	if len(p.CommanderDamage) > 0 {
		out.CommanderDamage = make(map[uuid.UUID]int, len(p.CommanderDamage))
		for k, v := range p.CommanderDamage {
			out.CommanderDamage[k] = v
		}
	} else {
		out.CommanderDamage = make(map[uuid.UUID]int)
	}
	if len(p.CommanderCasts) > 0 {
		out.CommanderCasts = make(map[uuid.UUID]int, len(p.CommanderCasts))
		for k, v := range p.CommanderCasts {
			out.CommanderCasts[k] = v
		}
	} else {
		out.CommanderCasts = make(map[uuid.UUID]int)
	}
	if len(p.Counters) > 0 {
		out.Counters = make(map[string]int, len(p.Counters))
		for k, v := range p.Counters {
			out.Counters[k] = v
		}
	}
	out.MaxHandSize = p.MaxHandSize
	if len(p.LifeHistory) > 0 {
		out.LifeHistory = make([]LifeChange, len(p.LifeHistory))
		copy(out.LifeHistory, p.LifeHistory)
	}
	return out
}

// cloneStackItem deep-copies a StackItem. Targets / Modes / Distribution
// are reallocated; scalar fields are value-copied. Returns nil for a
// nil input so the caller doesn't have to guard.
func cloneStackItem(s *StackItem) *StackItem {
	if s == nil {
		return nil
	}
	out := &StackItem{
		ID:           s.ID,
		Kind:         s.Kind,
		Controller:   s.Controller,
		Owner:        s.Owner,
		SourceCardID: s.SourceCardID,
		Label:        s.Label,
		XValue:       s.XValue,
		HoldPriority: s.HoldPriority,
		SplitSecond:  s.SplitSecond,
	}
	if len(s.Targets) > 0 {
		out.Targets = make([]TargetRef, len(s.Targets))
		copy(out.Targets, s.Targets)
	}
	if len(s.Modes) > 0 {
		out.Modes = make([]int, len(s.Modes))
		copy(out.Modes, s.Modes)
	}
	if len(s.Distribution) > 0 {
		out.Distribution = make(map[uuid.UUID]int, len(s.Distribution))
		for k, v := range s.Distribution {
			out.Distribution[k] = v
		}
	}
	return out
}

func cloneVote(v *Vote) *Vote {
	out := &Vote{
		ID:        v.ID,
		Topic:     v.Topic,
		Initiator: v.Initiator,
	}
	if len(v.Options) > 0 {
		out.Options = make([]string, len(v.Options))
		copy(out.Options, v.Options)
	}
	if len(v.Ballots) > 0 {
		out.Ballots = make(map[uuid.UUID]int, len(v.Ballots))
		for k, val := range v.Ballots {
			out.Ballots[k] = val
		}
	} else {
		out.Ballots = make(map[uuid.UUID]int)
	}
	return out
}

// RestoreFrom replaces this game's state with src while keeping the
// receiver's mutex untouched. Used by the room's undo stack: src is
// always a previously-captured Clone() that no other goroutine
// references, so a straight field copy is safe and the caller does
// not need to lock src.
//
// The receiver's mu is intentionally NOT touched — it stays bound to
// this *Game so existing hub / handler references continue to lock
// the same mutex they did before the restore.
//
// Caller MUST hold g.mu (write).
func (g *Game) RestoreFrom(src *Game) {
	g.ID = src.ID
	g.CreatedAt = src.CreatedAt
	g.State = src.State
	g.Seats = src.Seats
	g.Battlefield = src.Battlefield
	g.Stack = src.Stack
	g.Exile = src.Exile
	g.Turn = src.Turn
	g.MulligansOpen = src.MulligansOpen
	g.Monarch = src.Monarch
	g.Initiative = src.Initiative
	g.UndoLimit = src.UndoLimit
	g.StartingSeat = src.StartingSeat
	g.SplitSecondActive = src.SplitSecondActive
	g.StackMeta = src.StackMeta
	g.PendingTriggers = src.PendingTriggers
	g.LoyaltyActivatedThisTurn = src.LoyaltyActivatedThisTurn
	g.DiscardPending = src.DiscardPending
	g.Promises = src.Promises
	g.Vote = src.Vote
	g.rng = src.rng
}
