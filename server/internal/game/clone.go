package game

import "github.com/google/uuid"

// clone.go gives Game a deep-copy snapshot used by the undo stack
// and (later) replay machinery. The exported Snapshot already
// returns a shallow copy fine for read-only inspection; Clone goes
// further and gives back an entirely independent *Game whose
// mutation cannot affect the receiver.
//
// Cloned state intentionally drops the rwmutex (a fresh receiver
// owns its own) and COPIES the randomness: the RNG key and the
// per-stream draw counters (rng.go). Undo therefore REWINDS the
// random stream — restoring a clone and re-applying the same action
// draws exactly what the undone action drew, so undo cannot be used
// to reshuffle, re-roll or re-pick (ADR 0054 Decision 4, which
// reversed the old "undo does not rewind" contract).

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
		Settings:          g.Settings,
		StartingSeat:      g.StartingSeat,
		SplitSecondActive: g.SplitSecondActive,
		// #628: both halves of the CR 726 breaker. The threshold is
		// configuration and copies by value; the notice is a per-turn
		// fact an undo must be able to rewind past, so it gets its own
		// pointer rather than sharing the live one.
		LoopThreshold: g.LoopThreshold,
		LoopNotice:    cloneLoopNotice(g.LoopNotice),
		// The randomness is copied, not shared: the key by value
		// (an array) and the counters deeply, because randForLocked
		// increments the live map in place. The clone therefore
		// remembers how many draws each stream had taken, and
		// RestoreFrom puts that back — undo rewinds randomness, the
		// same way a snapshot restore resumes it (ADR 0054
		// Decisions 3-4).
		rngKey:            g.rngKey,
		rngCounters:       cloneRNGCounters(g.rngCounters),
		rngTurn:           g.rngTurn,
		sourceOrdinals:    cloneSourceOrdinals(g.sourceOrdinals),
		sourceOrdinalNext: g.sourceOrdinalNext,
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
	// S22 delayed triggers. Deep-copied for the same reason
	// PendingTriggers is: an undo that rewinds past the spell which
	// scheduled the return must un-schedule it, and one that rewinds
	// to a point where it was still owed must restore it intact.
	// The Effect func is shared — see cloneDelayedTrigger.
	if len(g.DelayedTriggers) > 0 {
		out.DelayedTriggers = make([]*DelayedTrigger, len(g.DelayedTriggers))
		for i, d := range g.DelayedTriggers {
			out.DelayedTriggers[i] = cloneDelayedTrigger(d)
		}
	}
	if len(g.LoyaltyActivatedThisTurn) > 0 {
		out.LoyaltyActivatedThisTurn = make(map[uuid.UUID]bool, len(g.LoyaltyActivatedThisTurn))
		for k, v := range g.LoyaltyActivatedThisTurn {
			out.LoyaltyActivatedThisTurn[k] = v
		}
	}
	if len(g.SpellsCastThisTurn) > 0 {
		out.SpellsCastThisTurn = make(map[uuid.UUID]CastTally, len(g.SpellsCastThisTurn))
		for k, v := range g.SpellsCastThisTurn {
			out.SpellsCastThisTurn[k] = v
		}
	}
	if len(g.ForetoldThisTurn) > 0 {
		out.ForetoldThisTurn = make(map[uuid.UUID]int, len(g.ForetoldThisTurn))
		for k, v := range g.ForetoldThisTurn {
			out.ForetoldThisTurn[k] = v
		}
	}
	if len(g.LandsPlayedThisTurn) > 0 {
		out.LandsPlayedThisTurn = make(map[uuid.UUID]int, len(g.LandsPlayedThisTurn))
		for k, v := range g.LandsPlayedThisTurn {
			out.LandsPlayedThisTurn[k] = v
		}
	}
	if len(g.ExtraLandDropsThisTurn) > 0 {
		out.ExtraLandDropsThisTurn = make(map[uuid.UUID]int, len(g.ExtraLandDropsThisTurn))
		for k, v := range g.ExtraLandDropsThisTurn {
			out.ExtraLandDropsThisTurn[k] = v
		}
	}
	out.TurnTally = cloneTurnTally(g.TurnTally)
	out.Activations = cloneActivationTally(g.Activations)
	if len(g.DrawnThisTurn) > 0 {
		out.DrawnThisTurn = make(map[uuid.UUID][]uuid.UUID, len(g.DrawnThisTurn))
		for k, v := range g.DrawnThisTurn {
			// Fresh backing array per player: the live map is appended
			// to on every draw, and an undo snapshot sharing it would
			// grow as the game it was taken from moved on.
			out.DrawnThisTurn[k] = append([]uuid.UUID(nil), v...)
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
	// #1199: the phased-out holding slice. Deep-copied like any other
	// zone, which is what makes an undo across a phase-out or a
	// phase-in exact — the state is a Card value in a slice, not a
	// flag several subsystems have to agree about. See ADR 0084.
	out.PhasedOut = cloneZone(g.PhasedOut)
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
	// PendingChoices: deep-copy pointers.
	if len(g.PendingChoices) > 0 {
		out.PendingChoices = make([]*PendingChoice, len(g.PendingChoices))
		for i, c := range g.PendingChoices {
			if c == nil {
				continue
			}
			cloned := *c
			// S15: ColorOptions is a slice — copy the backing array
			// so post-clone mutations on one don't leak to the other.
			if len(c.ColorOptions) > 0 {
				cloned.ColorOptions = append([]string(nil), c.ColorOptions...)
			}
			// S32 mana pipeline (#352): the spend restrictions a
			// PendingChoiceMana will stamp onto the token it mints.
			// New game state, so it needs its own backing array for
			// exactly the reason ColorOptions does — an undo that
			// shared it would let the restored game mutate the live
			// one, and the thing being shared here decides what the
			// mana may legally pay for.
			if len(c.ManaRestrictions) > 0 {
				cloned.ManaRestrictions = append([]string(nil), c.ManaRestrictions...)
			}
			// #742: the per-colour amounts of a one-pick-N-mana
			// choice. A map, so it needs its own copy for the same
			// reason the slices do.
			cloned.ManaAmounts = copyManaAmounts(c.ManaAmounts)
			if len(c.TriggerOrderIDs) > 0 {
				cloned.TriggerOrderIDs = append([]uuid.UUID(nil), c.TriggerOrderIDs...)
			}
			if len(c.PickTargetPlayers) > 0 {
				cloned.PickTargetPlayers = append([]uuid.UUID(nil), c.PickTargetPlayers...)
			}
			if len(c.PickTargetCards) > 0 {
				cloned.PickTargetCards = append([]uuid.UUID(nil), c.PickTargetCards...)
			}
			// #764 mode_pick: the offered options and their labels.
			// Own backing arrays for the same reason every other slice
			// here gets one.
			if len(c.ModeOptionIndex) > 0 {
				cloned.ModeOptionIndex = append([]int(nil), c.ModeOptionIndex...)
			}
			if len(c.ModeOptionLabel) > 0 {
				cloned.ModeOptionLabel = append([]string(nil), c.ModeOptionLabel...)
			}
			// S22 search chooser: the candidate list is a slice, so
			// it needs its own backing array for the same reason
			// every other slice here does — an undo that shared it
			// would let the restored game mutate the live one.
			if len(c.SearchCards) > 0 {
				cloned.SearchCards = append([]uuid.UUID(nil), c.SearchCards...)
			}
			// S16.5 copy chooser: same reason again — the candidate
			// list a Clone is choosing from is a slice.
			if len(c.CopyOptions) > 0 {
				cloned.CopyOptions = append([]uuid.UUID(nil), c.CopyOptions...)
			}
			// The chained-choice candidate list, for the same reason
			// as every slice above it.
			if len(c.ChooseCards) > 0 {
				cloned.ChooseCards = append([]uuid.UUID(nil), c.ChooseCards...)
			}
			// #568: the branches of an option pick. A slice of
			// structs each holding a card slice, so the deep copy
			// goes one level further than every other field here —
			// a shallow copy would share the pile arrays with the
			// undo snapshot.
			cloned.PickOptions = cloneChoiceOptions(c.PickOptions)
			// #793: the replacement resume frame holds the in-flight
			// ReplacementEvent, and answering the prompt MUTATES it —
			// Doubling Season doubles CounterDelta in place, Rhox
			// Faithmender doubles LifeDelta, and a life change's
			// continuation is consumed as it runs. Sharing that event
			// with the undo snapshot makes an UNDONE answer
			// unrepeatable: the replay doubles an already-doubled
			// value and finds the continuation gone. Giving the
			// snapshot its own copy is what makes "undo the answer,
			// answer again" land where answering once would.
			//
			// The gathered `applicable` list stays shared, like every
			// other server-only frame on a PendingChoice: a resume
			// reads it and never writes it.
			cloned.replacementResume = cloneReplacementResume(c.replacementResume)
			if c.coinFlipResume != nil {
				frame := *c.coinFlipResume
				cloned.coinFlipResume = &frame
			}
			out.PendingChoices[i] = &cloned
		}
	}
	// Event log: SHARED, not copied (#629).
	//
	// The log is append-only and 184 bytes an entry, so the deep copy
	// this used to be made every undo snapshot — one per action —
	// cost O(events) in time and memory. A few thousand events is
	// invisible; a long game or a trigger loop makes each action copy
	// megabytes, and the total is quadratic in the length of the
	// game.
	//
	// Sharing is safe because of what a snapshot does with the log:
	// it only ever reads entries that already existed when it was
	// taken, and entries never change after they are appended. The
	// recorded length is the slice header's own len, and RestoreFrom
	// truncating to it is what makes an undo drop exactly the events
	// the undone action emitted.
	//
	// The three-index slice is the safety belt. Capping cap to len
	// means a write through the snapshot — an append by some future
	// caller that decides to mutate a clone — reallocates instead of
	// scribbling into the live game's backing array past its length,
	// and equally that the live game's appends after a RestoreFrom
	// cannot rewrite entries an older snapshot still points at. The
	// cost is one reallocation on the first event emitted after an
	// undo, which is a rounding error against a copy per action.
	//
	// The persisted snapshot in snapshot.go is a different animal and
	// still copies: it serialises to JSON and outlives the process.
	//
	// Listeners are process-lifetime singletons — shallow-copy the
	// slice so the clone dispatches to the same subscribers the
	// original did.
	out.Events = g.Events[:len(g.Events):len(g.Events)]
	out.eventSeq = g.eventSeq
	// #829: the batch counter rewinds with the log, and the
	// once-per-batch marks rewind with it. Carrying one without the
	// other is the whole bug class: the counter alone would let an
	// undone trigger fire twice for one batch, the marks alone would
	// swallow the re-done one.
	out.eventBatch = g.eventBatch
	out.oncePerBatchFired = copyStringUint64Map(g.oncePerBatchFired)
	// #830: the block declaration's announcements rewind with the
	// declaration. An undo across a re-point that kept them would
	// swallow the re-done "becomes blocked"; dropping them would
	// announce the same attacker twice.
	out.announcedBlocks = copyUUIDPairMap(g.announcedBlocks)
	out.blockedAttackers = copyBoolMap(g.blockedAttackers)
	// #859: the attack declaration's announcements rewind with it for
	// the same reason — an undo across a re-point that kept them
	// would swallow the re-done attack trigger.
	out.announcedAttacks = copyBoolMap(g.announcedAttacks)
	// #716: and the combat damage steps' participation record rewinds
	// with the combat it belongs to. An undo back into the priority
	// window between the two steps that dropped it would let every
	// first-striker deal its damage a second time in the regular step.
	out.firstStrikeStepParticipants = copyBoolMap(g.firstStrikeStepParticipants)
	if len(g.Listeners) > 0 {
		out.Listeners = make([]Listener, len(g.Listeners))
		copy(out.Listeners, g.Listeners)
	}
	// Replacement registries (S17). Effects are immutable value
	// structs (func fields are process-lifetime) — copying the slice
	// headers into fresh backing arrays is enough; what matters is
	// that an append/clear on the original after the snapshot can't
	// reach the clone and vice versa.
	if len(g.BuiltinReplacements) > 0 {
		out.BuiltinReplacements = make([]ReplacementEffect, len(g.BuiltinReplacements))
		copy(out.BuiltinReplacements, g.BuiltinReplacements)
	}
	if len(g.TurnScopedReplacements) > 0 {
		out.TurnScopedReplacements = make([]ReplacementEffect, len(g.TurnScopedReplacements))
		copy(out.TurnScopedReplacements, g.TurnScopedReplacements)
	}
	// #750: the same reasoning for the until-end-of-turn block rules.
	// A BlockRule is written once at registration and never mutated,
	// so a fresh backing array is all the isolation an undo needs —
	// what must not be shared is the array, because the cleanup sweep
	// replaces the slice rather than compacting it.
	if len(g.TurnScopedBlockRules) > 0 {
		out.TurnScopedBlockRules = make([]BlockRule, len(g.TurnScopedBlockRules))
		copy(out.TurnScopedBlockRules, g.TurnScopedBlockRules)
	}
	// S32/S38 scoped statics — the layer-engine twin of the slice
	// above, and the same reasoning: a ScopedStatic is written once
	// at registration and never mutated (see the immutability
	// contract on the type), so a fresh backing array is enough.
	// What must not be shared is the array itself — the cleanup-step
	// sweeps replace the slice rather than compacting in place
	// precisely so an undo snapshot taken mid-turn still holds the
	// grants that were live when it was taken.
	if len(g.ScopedStatics) > 0 {
		out.ScopedStatics = make([]ScopedStatic, len(g.ScopedStatics))
		copy(out.ScopedStatics, g.ScopedStatics)
	}
	// CR 603.10 LKI snapshots (S19). Values are Characteristic copies
	// that are never mutated after being stored, so a per-entry value
	// copy is sufficient. Usually empty — entries live only for the
	// duration of one LTB-emitting mutation.
	if len(g.lastKnownBattlefield) > 0 {
		out.lastKnownBattlefield = make(map[uuid.UUID]Characteristic, len(g.lastKnownBattlefield))
		for k, v := range g.lastKnownBattlefield {
			out.lastKnownBattlefield[k] = v
		}
	}
	if len(g.lastKnownTriggerIdentity) > 0 {
		out.lastKnownTriggerIdentity = make(map[uuid.UUID]triggerIdentityLKI, len(g.lastKnownTriggerIdentity))
		for k, v := range g.lastKnownTriggerIdentity {
			out.lastKnownTriggerIdentity[k] = v
		}
	}
	// #1218: lastKnownCounters' values are themselves maps, so unlike
	// its two siblings above each entry needs its OWN copy — sharing
	// the inner map would let a mutation on the clone (or the undo it
	// is taken for) write through to the live game's snapshot.
	if len(g.lastKnownCounters) > 0 {
		out.lastKnownCounters = make(map[uuid.UUID]map[string]int, len(g.lastKnownCounters))
		for k, v := range g.lastKnownCounters {
			out.lastKnownCounters[k] = copyStringIntMap(v)
		}
	}
	// #1255: an undo across a counterspell rewinds the record with the
	// spell, so a replay that counters it again records it again.
	out.lastKnownStack = cloneLastKnownStack(g.lastKnownStack)
	// CR 614.5 once-per-event marks for events PAUSED on a CR 616 /
	// CR 614.10 prompt (#808). Between actions the map holds an entry
	// only for an event whose prompt is still open, and that entry is
	// part of the prompt's state: an effect that applied on its own
	// before the prompt was queued is marked there, and nowhere else.
	// Answering the prompt deletes the entry when the event settles, so
	// an undo snapshot that did not carry its own copy would replay the
	// answer with the mark gone — and the effect would fire a second
	// time. Deep-copied, because the live pipeline writes the inner
	// sets in place.
	if len(g.enteringTokens) > 0 {
		// #762: a token whose entry paused on a replacement prompt is
		// here and nowhere else. An undo across that prompt has to
		// bring it back, exactly as it brings back the once-per-event
		// marks the same paused window is holding.
		out.enteringTokens = append([]Card(nil), g.enteringTokens...)
	}
	// #920: the item currently resolving, for the same reason — a
	// spell copying ITSELF makes the copy from a prompt's answer, and
	// an undo across that prompt has to hand the restored game the
	// same source. The POINTER is shared rather than deep-copied,
	// deliberately: the paused prompt's own resume frame holds that
	// same *StackItem (may_choice.go's contract), so a second copy
	// would be a second item the frame is not writing through. The
	// struct is replaced wholesale at each event batch and never
	// mutated in place, so nothing can diverge.
	out.resolving = g.resolving
	if len(g.replacementsAppliedThisEvent) > 0 {
		out.replacementsAppliedThisEvent = make(map[ReplacementEventID]map[ReplacementEffectID]bool, len(g.replacementsAppliedThisEvent))
		for evID, set := range g.replacementsAppliedThisEvent {
			cp := make(map[ReplacementEffectID]bool, len(set))
			for k, v := range set {
				cp[k] = v
			}
			out.replacementsAppliedThisEvent[evID] = cp
		}
	}
	// #1019, #1027: the prompted runs in flight — sacrifices and
	// discards alike. Deep in the counter and the per-seat landed
	// lists, shallow in the continuation closure — clonePromptRuns
	// says why, and it is the same split cloneReplacementResume makes.
	out.promptRuns = clonePromptRuns(g.promptRuns)
	// S16 layer-engine version counters. Atomics can't be struct-
	// copied; mirror via Load/Store so the clone's staleness state
	// matches the original's at capture time.
	out.layerVersion.Store(g.layerVersion.Load())
	out.lastResolvedVersion.Store(g.lastResolvedVersion.Load())
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
	if len(c.Keywords) > 0 {
		out.Keywords = append([]string(nil), c.Keywords...)
	}
	if len(c.GrantedAbilities) > 0 {
		out.GrantedAbilities = append([]string(nil), c.GrantedAbilities...)
	}
	if len(c.ManaAbilities) > 0 {
		out.ManaAbilities = append([]ManaAbilityShape(nil), c.ManaAbilities...)
	}
	if len(c.ActivatedAbilities) > 0 {
		out.ActivatedAbilities = append([]ActivatedAbilityShape(nil), c.ActivatedAbilities...)
	}
	if len(c.Colors) > 0 {
		out.Colors = append([]string(nil), c.Colors...)
	}
	// ADR 0034 faces. ActiveFace and Layout are scalars and ride the
	// value copy, but Faces is a slice of structs each holding its
	// own Colors slice — two levels of aliasing, both of which would
	// survive an undo. The face list itself is immutable printed data
	// today (only SetFace reads it), so this is belt-and-braces; a
	// later transform effect that rewrites a face in place would make
	// it load-bearing, and by then the aliasing bug would be silent.
	if len(c.Faces) > 0 {
		out.Faces = make([]Face, len(c.Faces))
		copy(out.Faces, c.Faces)
		for i := range out.Faces {
			if len(c.Faces[i].Colors) > 0 {
				out.Faces[i].Colors = append([]string(nil), c.Faces[i].Colors...)
			}
		}
	}
	// S16.5 copy effects: PrintedSelf is a POINTER, so the value copy
	// aliases it between the live game and every undo snapshot. A
	// Clone that dies after the snapshot would restore its printed
	// values through the shared pointer and the undo would find it
	// already un-cloned.
	out.PrintedSelf = copyPrintedValues(c.PrintedSelf)
	// #1270: the listed face-down body is a pointer too. Never
	// mutated through, but copied for PrintedSelf's reason — an undo
	// snapshot must not share anything with the live card.
	out.FaceDownListed = c.FaceDownListed.clone()
	if len(c.Counters) > 0 {
		out.Counters = make(map[string]int, len(c.Counters))
		for k, v := range c.Counters {
			out.Counters[k] = v
		}
	} else {
		out.Counters = nil
	}
	if len(c.NextUntapSkips) > 0 {
		out.NextUntapSkips = append([]UntapSkip(nil), c.NextUntapSkips...)
	} else {
		out.NextUntapSkips = nil
	}
	// CR 400.7d (#653 / ADR 0073): what the spell that became this
	// permanent was cast for. Its one slice would alias the live
	// record into every undo snapshot under a value copy.
	out.Provenance = c.Provenance.Clone()
	// S13.5 knowledge set: a value copy would alias the live map, so
	// reveals after the snapshot would leak into it and undo couldn't
	// roll knowledge back.
	if len(c.KnownBy) > 0 {
		out.KnownBy = make(map[uuid.UUID]bool, len(c.KnownBy))
		for k, v := range c.KnownBy {
			out.KnownBy[k] = v
		}
	} else {
		out.KnownBy = nil
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
		TurnsBegun:        p.TurnsBegun,
		Eliminated:        p.Eliminated,
		HandKept:          p.HandKept,
		MulligansTaken:    p.MulligansTaken,
		DeckImported:      p.DeckImported,
		UndosRemaining:    p.UndosRemaining,
		DiscordID:         p.DiscordID,
		DiscordAvatarHash: p.DiscordAvatarHash,
		DisplayName:       p.DisplayName,
		IsBot:             p.IsBot,
		BotTier:           p.BotTier,
		BotDeck:           p.BotDeck,
	}
	out.Library = cloneZone(p.Library)
	out.Hand = cloneZone(p.Hand)
	out.Graveyard = cloneZone(p.Graveyard)
	out.Command = cloneZone(p.Command)
	out.Emblems = cloneZone(p.Emblems)
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
	out.LandDropsPerTurn = p.LandDropsPerTurn
	if len(p.LifeHistory) > 0 {
		out.LifeHistory = make([]LifeChange, len(p.LifeHistory))
		copy(out.LifeHistory, p.LifeHistory)
	}
	// S15: deep-copy ManaPool so undo restores the exact token
	// identity (Source uuid + Restrictions slice) rather than
	// aliasing the originals.
	if len(p.ManaPool) > 0 {
		out.ManaPool = make(ManaPool, len(p.ManaPool))
		for i, t := range p.ManaPool {
			cloned := t
			if len(t.Restrictions) > 0 {
				cloned.Restrictions = append([]string(nil), t.Restrictions...)
			}
			out.ManaPool[i] = cloned
		}
	}
	// ADR 0066: granted cast and play permissions. Deep-copied for the
	// reason the mana pool is — an undo that rewinds past the
	// Snapcaster must take the flashback back, and one that rewinds to
	// a point where the permission was still owed must restore it
	// intact, with its own Cards backing array.
	out.CastPermissions = cloneCastPermissions(p.CastPermissions)
	// #1197: granted player abilities, the same reasoning one more
	// time. A PlayerStatic has no reference-typed field at all, so a
	// fresh backing array IS the whole copy — what must not be shared
	// is the array, because sweepPlayerStaticsLocked replaces the
	// slice rather than compacting it, precisely so an undo snapshot
	// taken mid-turn still holds the grants that were live then.
	out.Statics = clonePlayerStatics(p.Statics)
	return out
}

// cloneCastPermissions deep-copies a player's granted permissions.
// The only reference-typed fields are Cards and Faces, so two
// reallocations per permission are the whole copy; everything else is
// scalar, which is exactly the property that lets the snapshot mirror
// the type rather than rebuild it.
func cloneCastPermissions(in []CastPermission) []CastPermission {
	if len(in) == 0 {
		return nil
	}
	out := make([]CastPermission, len(in))
	copy(out, in)
	for i := range out {
		if len(in[i].Cards) > 0 {
			out[i].Cards = append([]PermissionCardRef(nil), in[i].Cards...)
		}
		if len(in[i].Faces) > 0 {
			out[i].Faces = append([]int(nil), in[i].Faces...)
		}
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
		ID:            s.ID,
		Kind:          s.Kind,
		Controller:    s.Controller,
		Owner:         s.Owner,
		SourceCardID:  s.SourceCardID,
		SourceEpoch:   s.SourceEpoch,
		Label:         s.Label,
		DoubledBy:     s.DoubledBy,
		DoubledByName: s.DoubledByName,
		XValue:        s.XValue,
		HoldPriority:  s.HoldPriority,
		SplitSecond:   s.SplitSecond,
		AltCost:       s.AltCost,
		Foretold:      s.Foretold,
		FaceDown:      s.FaceDown,
		AltCostExiles: s.AltCostExiles,
		CastFromZone:  s.CastFromZone,
		IsCopy:        s.IsCopy,
		Seq:           s.Seq,
		// #789 / #761: what the announcement paid. Deep-copied
		// (clonePaidCost reallocates the token slice and each
		// token's Restrictions) because an undo snapshot that
		// aliased the live backing array would let a restore mutate
		// the game it was taken from.
		Paid: clonePaidCost(s.Paid),
		// Effect takes the live *Game at resolve time rather than
		// capturing one, so sharing the func between original and
		// snapshot is safe — an undo that restores this item
		// resolves it against the restored game.
		Effect:     s.Effect,
		Ordered:    s.Ordered,
		targetSpec: s.targetSpec,
		// #764: catalog data, read-never-written, so the undo clone
		// shares the pointer exactly as it shares targetSpec.
		modeSpec: s.modeSpec,
	}
	if len(s.Targets) > 0 {
		out.Targets = make([]TargetRef, len(s.Targets))
		copy(out.Targets, s.Targets)
	}
	// #1223: the triggering event, deep-copied for the same reason
	// the payload is — an undo that shared the slices inside it
	// would let the restored game mutate the live one.
	out.Trigger = cloneTriggerContext(s.Trigger)
	// A reflexive trigger's payload (#636): its own backing array for
	// the same reason Targets gets one — an undo that shared it would
	// let the restored game mutate the live one.
	if len(s.Payload) > 0 {
		out.Payload = make([]TargetRef, len(s.Payload))
		copy(out.Payload, s.Payload)
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

// cloneReplacementResume gives an undo snapshot its own copy of the
// in-flight ReplacementEvent a paused CR 614 pipeline is sitting on.
//
// Only the EVENT is copied. The gathered `applicable` list is shared:
// a resume reads it to decide what to fire and never writes it. Its
// entries' `source` pointers point into the LIVE battlefield the gather
// walked, not into the snapshot's copy of it — harmless, because the
// resume only ever reads through them — which is the same shallow
// sharing every other server-only resume frame on a PendingChoice
// already has.
//
// The event's once-per-event marks (CR 614.5) are not on the frame:
// they live in Game.replacementsAppliedThisEvent, which cloneLocked
// deep-copies and RestoreFrom puts back, so a replayed answer skips
// exactly the effects the first answer skipped (#808).
//
// What the resume writes is the event's scalar payload — the counter
// delta, the life delta, Canceled — and the lifeTail and counterTail
// POINTERS, which a continuation clears on the event as it runs. Both
// live in the struct this copies, so the snapshot keeps the values the
// prompt was queued with. #793, #1282.
//
// The damageTail and the zoneRoute are the exceptions, and it is why
// they each get a copy of their own (#807, #853). Their continuations
// are cleared THROUGH the pointer — runDamageTailLocked and
// runRouteTailLocked nil `then` on the tail rather than the tail on the
// event, because the REST of what they carry (the CR 120.3 target
// kind, the deathtouch / lifelink / commander snapshot; the
// destination, the to-the-bottom instruction, the discard flag) is
// what applyResolvedDamageLocked and executeZoneRouteLocked are still
// reading when it runs. Sharing the struct would let the live game's
// run consume the snapshot's continuation, so undoing the answer to a
// CR 903.9 prompt and answering it again would move the card and skip
// the rest of the discard.
func cloneReplacementResume(f *replacementResumeFrame) *replacementResumeFrame {
	if f == nil {
		return nil
	}
	out := *f
	if f.ev != nil {
		ev := *f.ev
		if len(f.ev.EntersWithCounters) > 0 {
			ev.EntersWithCounters = make(map[string]int, len(f.ev.EntersWithCounters))
			for k, v := range f.ev.EntersWithCounters {
				ev.EntersWithCounters[k] = v
			}
		}
		if f.ev.damageTail != nil {
			t := *f.ev.damageTail
			ev.damageTail = &t
		}
		if f.ev.tokenTail != nil {
			// #762, the same reason as the two below: the token tail is
			// cleared THROUGH the pointer as it runs, so sharing it
			// would let the live game's run consume the snapshot's
			// continuation and an undone-then-redone answer would make
			// the tokens and skip the rest of the card.
			t := *f.ev.tokenTail
			ev.tokenTail = &t
		}
		if len(f.ev.TokenGroups) > 0 {
			// The settled creation itself — how many of which kinds —
			// is what the resume is holding. Its own slice, so a
			// doubler applied on the live game cannot reach back into
			// the snapshot's copy of the event.
			ev.TokenGroups = append([]TokenGroup(nil), f.ev.TokenGroups...)
		}
		if f.ev.keywordAction != nil {
			// #976, the same reason as the token tail above: the
			// keyword action's continuation is cleared THROUGH the
			// pointer when the action is abandoned, and the REST of
			// what the struct carries — the proliferate's chosen
			// lists, the prompt kind a settled scry queues — is what
			// applyResolvedKeywordActionLocked is still reading. Its
			// own copy, with its own slices, so a live run cannot
			// consume the snapshot's continuation or scribble into the
			// choice an undone answer would replay.
			t := *f.ev.keywordAction
			t.cards = append([]uuid.UUID(nil), f.ev.keywordAction.cards...)
			t.players = append([]uuid.UUID(nil), f.ev.keywordAction.players...)
			// #1236: the amass token template is a Card, and a Card
			// carries slices (Colors, Keywords) the minting path
			// appends to. cloneCard for the same reason the two
			// slices above get their own backing arrays.
			t.armyToken = cloneCard(f.ev.keywordAction.armyToken)
			ev.keywordAction = &t
		}
		if f.ev.mill != nil {
			// #569, the same reason as the two above: the mill's
			// continuation is cleared THROUGH the pointer as it runs,
			// and the REST of what the tail carries — the destination —
			// is what applyResolvedMillLocked is still reading. Its own
			// copy, so a live run cannot consume the snapshot's
			// continuation and an undone-then-redone answer mills the
			// same cards and reports the same list.
			t := *f.ev.mill
			ev.mill = &t
		}
		if f.ev.zoneRoute != nil {
			r := *f.ev.zoneRoute
			if len(f.ev.zoneRoute.simultaneousExit) > 0 {
				r.simultaneousExit = append([]Card(nil), f.ev.zoneRoute.simultaneousExit...)
			}
			ev.zoneRoute = &r
		}
		// #478: and the ENTRY tail, for exactly the reason the route
		// gets a copy. runEntryTailLocked nils `then` on the tail rather
		// than the tail on the event, because the rest of what it
		// carries (the CR 400.7 new-object flag) is what
		// executeEntryToBattlefieldLocked is still reading when it runs.
		// Sharing the struct would let the live game's run consume the
		// snapshot's continuation, so undoing the answer to a fetched
		// permanent's entry prompt and answering it again would push the
		// card and skip the library shuffle.
		if f.ev.entryTail != nil {
			tail := *f.ev.entryTail
			// #1322: and the simultaneous entry the tail may be one
			// card of. The batch's cursor moves and its settled events
			// accumulate as the resume walks it, so a shared batch
			// would let the live game's answer walk the snapshot's
			// batch on, and an undone-then-redone answer would skip
			// the rest of the entry.
			tail.batch = cloneEntryBatch(f.ev.entryTail.batch)
			ev.entryTail = &tail
		}
		out.ev = &ev
	}
	return &out
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
	g.PhasedOut = src.PhasedOut
	g.Turn = src.Turn
	g.MulligansOpen = src.MulligansOpen
	g.Monarch = src.Monarch
	g.Initiative = src.Initiative
	// Settings are NOT restored (ADR 0075 §2.3): the live value is
	// carried forward, so an undo cannot roll back a settings change —
	// least of all the undo limit it is spending against. A snapshot
	// restore is a different path (restoreGame) and does restore them.
	g.StartingSeat = src.StartingSeat
	g.SplitSecondActive = src.SplitSecondActive
	g.StackMeta = src.StackMeta
	g.PendingTriggers = src.PendingTriggers
	g.DelayedTriggers = src.DelayedTriggers
	g.LoyaltyActivatedThisTurn = src.LoyaltyActivatedThisTurn
	g.SpellsCastThisTurn = src.SpellsCastThisTurn
	g.ForetoldThisTurn = src.ForetoldThisTurn
	g.LandsPlayedThisTurn = src.LandsPlayedThisTurn
	g.ExtraLandDropsThisTurn = src.ExtraLandDropsThisTurn
	g.DrawnThisTurn = src.DrawnThisTurn
	g.TurnTally = src.TurnTally
	// #1181: the activation record rewinds with the rest of the
	// per-turn state. An undo that kept an exhaust spent would take
	// the ability away for the whole game on the strength of an
	// activation that no longer happened.
	g.Activations = src.Activations
	// #628's loop breaker, missed by this list when it landed: the
	// clone carries the notice (cloneLocked, above) and the persisted
	// snapshot carries it, but the undo path did not put it back, so
	// an undo across the moment the breaker fired left the live notice
	// exactly as it was — stale in one direction or absent in the
	// other. #804 makes that visible rather than merely wrong: the
	// CR 726 prompt rewinds with PendingChoices, and a prompt without
	// the notice it is asking about is a question about nothing.
	g.LoopNotice = src.LoopNotice
	g.LoopThreshold = src.LoopThreshold
	g.DiscardPending = src.DiscardPending
	g.Promises = src.Promises
	g.Vote = src.Vote
	// The event log is TRUNCATED to the snapshot's length, not
	// copied back into place: src.Events is the same backing array
	// this game has been appending to, capped at the length it had
	// when the snapshot was taken (#629, see cloneLocked). Assigning
	// it drops exactly the events the undone action emitted, and
	// eventSeq rewinds with it so the next event carries the Seq the
	// undone one did. TurnTally.FirstEvent is an index into this log
	// and comes from the same snapshot, so it cannot point past the
	// restored end.
	g.Events = src.Events
	g.eventSeq = src.eventSeq
	// #829: batch identity rewinds with the log it is stamped into,
	// and the once-per-batch marks rewind with the counter — see
	// cloneLocked.
	g.eventBatch = src.eventBatch
	g.oncePerBatchFired = src.oncePerBatchFired
	// #830 / #859: see cloneLocked — the announcements rewind with
	// the declarations they describe.
	g.announcedBlocks = src.announcedBlocks
	g.blockedAttackers = src.blockedAttackers
	g.announcedAttacks = src.announcedAttacks
	g.firstStrikeStepParticipants = src.firstStrikeStepParticipants
	g.Listeners = src.Listeners
	g.PendingChoices = src.PendingChoices
	g.BuiltinReplacements = src.BuiltinReplacements
	g.TurnScopedReplacements = src.TurnScopedReplacements
	g.TurnScopedBlockRules = src.TurnScopedBlockRules
	g.ScopedStatics = src.ScopedStatics
	g.lastKnownBattlefield = src.lastKnownBattlefield
	g.lastKnownTriggerIdentity = src.lastKnownTriggerIdentity
	g.lastKnownCounters = src.lastKnownCounters
	// Copied, not shared: rememberLeavingSpellLocked inserts into the
	// live map, and an undo snapshot may be restored more than once.
	g.lastKnownStack = cloneLastKnownStack(src.lastKnownStack)
	// #808: the paused events' once-per-event marks rewind with the
	// prompts that own them — see cloneLocked.
	g.replacementsAppliedThisEvent = src.replacementsAppliedThisEvent
	// #1019, #1027: a half-answered prompted run rewinds with the
	// prompts still owing it — see cloneLocked.
	g.promptRuns = src.promptRuns
	g.enteringTokens = src.enteringTokens
	// #920: the resolving item rewinds with the prompt that is reading
	// it — see cloneLocked.
	g.resolving = src.resolving
	// The randomness rewinds with everything else: the key, the
	// per-stream draw counters and the turn they belong to (ADR 0054
	// Decision 4). Adopted like the other fields — src is consumed.
	g.rngKey = src.rngKey
	g.rngCounters = src.rngCounters
	g.rngTurn = src.rngTurn
	g.sourceOrdinals = src.sourceOrdinals
	g.sourceOrdinalNext = src.sourceOrdinalNext
	// S16 layer-engine counters: adopt the snapshot's values via
	// Store/Load (atomics can't be field-copied), then bump
	// layerVersion past lastResolvedVersion so the next snapshot
	// path runs a full recompute against the restored battlefield —
	// the restored cards' `effective` caches were computed for the
	// snapshot-time state and must not be served as-is.
	g.lastResolvedVersion.Store(src.lastResolvedVersion.Load())
	g.layerVersion.Store(src.layerVersion.Load() + 1)
}

func cloneSourceOrdinals(in map[uuid.UUID]uint64) map[uuid.UUID]uint64 {
	if len(in) == 0 {
		return nil
	}
	out := make(map[uuid.UUID]uint64, len(in))
	for id, ordinal := range in {
		out[id] = ordinal
	}
	return out
}
