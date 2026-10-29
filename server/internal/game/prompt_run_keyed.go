package game

import (
	"fmt"
	"sort"

	"github.com/google/uuid"
)

// prompt_run_keyed.go — a prompted run whose continuation is a KEY
// rather than a closure (#4290, ADR 0013 §5x / §5y amendment of
// 2026-10-24), so a table waiting on any leg of a multi-seat instruction
// is a restore point.
//
// A run (prompt_run.go) is {outstanding, seats, landed, then}, and
// everything but `then` was data already. Here `then` is a registered
// function, looked up by its key when the last leg settles, and what it
// needs travels on the run as plain data: the card that asked (Source)
// and a list of IDs the card put there (Carry). The run is written to a
// restore point as GameSnapshot.PromptRuns, and each leg names its run
// (PendingChoice.promptRun) on disk too.
//
// The key shares the effect-key namespace and is ledgered as
// "run <key>". A binary that does not know it refuses the file
// (checkEffectKeys). An older binary, which knows no run on disk at all,
// refuses it through the legs: a keyed discard leg names
// choose-cards/discard-leg and a keyed sacrifice leg names its run's key
// in PickThen, neither of which it has.
//
// Each verb keeps its own vocabulary over the one answer: a discard
// run's continuation reads RunDone.Discards(), a sacrifice run's
// RunDone.Sacrifices(), a pick run's RunDone.Picks().

// RunDone is what a keyed run's continuation is handed.
type RunDone struct {
	// Source is the card that asked.
	Source uuid.UUID
	// Landed is one entry per seat the run ASKED, in ask order, with
	// what that seat really moved (SeatCards). Nil when nobody was asked.
	Landed []SeatCards
	// Carry and Nums are the plain data the queuing card put on the run.
	Carry []uuid.UUID
	Nums  []int
	// Asked is the stack item that asked, as data (NewAskedItem), or nil
	// when the card put none there. Item rebuilds it.
	Asked *AskedItem
}

// Item is the stack item the run's continuation runs against: a copy
// rebuilt from the data the run carried (AskedItem.StackItem), nil when
// the card put none there. A new copy each call.
func (r RunDone) Item() *StackItem { return r.Asked.StackItem() }

// RunData is the plain data a keyed run hands its continuation beside the
// answer: IDs and numbers the card put there, and the stack item that asked
// (NewAskedItem). A run holds no closure, so everything the rest of the
// card read from the frame of its Go function travels here.
type RunData struct {
	Carry []uuid.UUID
	Nums  []int
	Asked *AskedItem
}

// Discards reads the answer as a discard run's (PromptedDiscards).
func (r RunDone) Discards() PromptedDiscards { return PromptedDiscards(r.Landed) }

// Sacrifices reads the answer as a sacrifice run's (PromptedSacrifices).
func (r RunDone) Sacrifices() PromptedSacrifices { return PromptedSacrifices(r.Landed) }

// Picks reads the answer as a pick run's (PromptedPicks).
func (r RunDone) Picks() PromptedPicks { return PromptedPicks(r.Landed) }

// RunLeg is what a keyed discard run's per-seat tail is handed: the
// seat whose discard has just finished, what it really discarded, and
// the run's plain data.
type RunLeg struct {
	Seat, Source uuid.UUID
	Landed       []uuid.UUID
	Carry        []uuid.UUID
}

// PromptRunFunc is a registered run continuation. It runs with g.mu
// held, may queue prompts of its own, and must capture nothing.
type PromptRunFunc func(g *Game, r RunDone) error

// PromptRunLegFunc is a registered per-seat tail: the rest of ONE
// seat's sentence ("…discards a card, then mills a card"), run as that
// seat's leg finishes and before the leg settles. Same rules as
// PromptRunFunc.
type PromptRunLegFunc func(g *Game, leg RunLeg) error

// PromptRunThen names a registered run continuation. The key is
// unexported, so the only way to hold one is RegisterPromptRunThen.
type PromptRunThen struct{ key string }

// Key is the continuation's on-disk key.
func (r PromptRunThen) Key() string { return r.key }

// Valid reports whether this names a registered continuation: false for
// the zero value.
func (r PromptRunThen) Valid() bool { return r.key != "" }

// promptRunThens and promptRunLegs are guarded by effectRegistryMu, with
// the other effect keys, because the keys share one namespace and one
// ledger. A leg tail shares its run's key and needs no ledger line.
var (
	promptRunThens = map[string]PromptRunFunc{}
	promptRunLegs  = map[string]PromptRunLegFunc{}
	promptRunRules = map[string]SetRuleFunc{}
)

// RegisterPromptRunThen registers a keyed run continuation. Call it
// once, from a package-level var. Panics on a bad or duplicate key. The
// key is an on-disk identity: never renamed, never reused, and ledgered
// as "run <key>".
func RegisterPromptRunThen(key string, fn PromptRunFunc) PromptRunThen {
	return RegisterPromptRunThenLeg(key, nil, fn)
}

// RegisterPromptRunThenLeg is RegisterPromptRunThen for a discard run
// whose seats each have the rest of their own sentence (Vicious
// Rumors' "then mills a card"). `leg` runs for each seat as its discard
// lands, and for a seat asked nothing because its hand was empty (CR
// 701.9a), as a discard's "then" always has (#797).
func RegisterPromptRunThenLeg(key string, leg PromptRunLegFunc, fn PromptRunFunc) PromptRunThen {
	effectRegistryMu.Lock()
	defer effectRegistryMu.Unlock()
	checkEffectKey("prompted-run continuation", key)
	if fn == nil {
		panic(fmt.Sprintf("game: prompted-run continuation %q has no function", key))
	}
	promptRunThens[key] = fn
	if leg != nil {
		promptRunLegs[key] = leg
	}
	return PromptRunThen{key: key}
}

// RegisterPromptRunThenRule is RegisterPromptRunThen for a run whose legs
// are permanent picks with a set rule (Tragic Arrogance's "an artifact, a
// creature, an enchantment and a planeswalker"): every answer to a leg,
// and every answer the enumerator offers, is checked against `rule` (nil
// allows every set the bounds allow). The rule shares the run's key, so a
// binary that does not know the key refuses the file and one that knows it
// has the rule. It is never called for an empty pick and is a function of
// the cards alone, so it must capture nothing.
func RegisterPromptRunThenRule(key string, rule SetRuleFunc, fn PromptRunFunc) PromptRunThen {
	then := RegisterPromptRunThen(key, fn)
	if rule != nil {
		effectRegistryMu.Lock()
		promptRunRules[key] = rule
		effectRegistryMu.Unlock()
	}
	return then
}

// lookupPromptRunRule is the set rule registered with a run, nil when it
// has none.
func lookupPromptRunRule(key string) SetRuleFunc {
	effectRegistryMu.RLock()
	defer effectRegistryMu.RUnlock()
	return promptRunRules[key]
}

func lookupPromptRunThen(key string) (PromptRunFunc, bool) {
	effectRegistryMu.RLock()
	defer effectRegistryMu.RUnlock()
	fn, ok := promptRunThens[key]
	return fn, ok
}

func lookupPromptRunLeg(key string) PromptRunLegFunc {
	effectRegistryMu.RLock()
	defer effectRegistryMu.RUnlock()
	return promptRunLegs[key]
}

// KnownPromptRunThen reports whether this binary has a run continuation
// registered under key.
func KnownPromptRunThen(key string) bool { _, ok := lookupPromptRunThen(key); return ok }

// registeredPromptRunKeysLocked lists the ledger lines, "run <key>".
// Caller holds effectRegistryMu.
func registeredPromptRunKeysLocked() []string {
	out := make([]string, 0, len(promptRunThens))
	for k := range promptRunThens {
		out = append(out, "run "+k)
	}
	sort.Strings(out)
	return out
}

// runLegTailLocked runs a keyed discard run's per-seat tail, if its
// continuation registered one, for `seat` with what it discarded. An
// error is the table's to see and does not stop the leg settling, as
// the closure form's leg Then does not.
//
// Caller must hold g.mu.
func (g *Game) runLegTailLocked(key string, seat, source uuid.UUID, landed, carry []uuid.UUID) {
	if key == "" {
		return
	}
	leg := lookupPromptRunLeg(key)
	if leg == nil {
		return
	}
	if err := leg(g, RunLeg{Seat: seat, Source: source, Landed: copyUUIDs(landed), Carry: copyUUIDs(carry)}); err != nil {
		g.emitChoiceEffectErrorLocked(seat, source, err)
	}
}

// promptRunSnapshot is one keyed run in flight on disk (#4290). A run
// whose continuation is a closure is not written: the census counts it
// instead (captureSnapshotLocked).
type promptRunSnapshot struct {
	ID          uuid.UUID `json:"id"`
	Outstanding int       `json:"outstanding"`
	// Seats is the ask order; Landed what each seat has moved so far,
	// for the seats that have moved anything, in that order.
	Seats  []uuid.UUID            `json:"seats,omitempty"`
	Landed []promptRunSeatLanding `json:"landed,omitempty"`
	Key    string                 `json:"key"`
	Carry  []uuid.UUID            `json:"carry,omitempty"`
	Source uuid.UUID              `json:"source,omitempty"`
	// Nums and Asked are RunData's other two fields. Additive within the
	// version: a file without them restores a run that reads none.
	Nums  []int      `json:"nums,omitempty"`
	Asked *AskedItem `json:"asked,omitempty"`
}

// promptRunSeatLanding is one seat's landed list on disk.
type promptRunSeatLanding struct {
	Seat  uuid.UUID   `json:"seat"`
	Cards []uuid.UUID `json:"cards"`
}

// snapshotPromptRuns writes the keyed runs in flight, ordered by ID so
// two captures of one table are byte-identical. Every run kept is keyed:
// the closure form, which the census counted here, is gone (#4361).
func snapshotPromptRuns(runs map[uuid.UUID]*promptRun) []promptRunSnapshot {
	if len(runs) == 0 {
		return nil
	}
	ids := make([]uuid.UUID, 0, len(runs))
	for id := range runs {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i].String() < ids[j].String() })
	var out []promptRunSnapshot
	for _, id := range ids {
		run := runs[id]
		if run == nil {
			continue
		}
		if run.thenKey == "" {
			// Never kept (startRunLocked): a run nobody waits on has no
			// link. The closure form is gone (#4361).
			continue
		}
		rs := promptRunSnapshot{
			ID:          id,
			Outstanding: run.outstanding,
			Seats:       copyUUIDs(run.seats),
			Key:         run.thenKey,
			Carry:       copyUUIDs(run.carry),
			Source:      run.source,
			Nums:        copyInts(run.nums),
			Asked:       copyAskedItem(run.asked),
		}
		for _, seat := range run.seats {
			if cards := run.landed[seat]; len(cards) > 0 {
				rs.Landed = append(rs.Landed, promptRunSeatLanding{Seat: seat, Cards: copyUUIDs(cards)})
			}
		}
		out = append(out, rs)
	}
	return out
}

// restorePromptRuns rebuilds the keyed runs. A file written before them
// restores none, which is the old behaviour (no keyed run existed).
func restorePromptRuns(in []promptRunSnapshot) map[uuid.UUID]*promptRun {
	if len(in) == 0 {
		return nil
	}
	out := make(map[uuid.UUID]*promptRun, len(in))
	for _, rs := range in {
		if rs.ID == uuid.Nil || rs.Key == "" {
			continue
		}
		run := &promptRun{
			outstanding: rs.Outstanding,
			seats:       copyUUIDs(rs.Seats),
			landed:      make(map[uuid.UUID][]uuid.UUID, len(rs.Landed)),
			thenKey:     rs.Key,
			carry:       copyUUIDs(rs.Carry),
			source:      rs.Source,
			nums:        copyInts(rs.Nums),
			asked:       copyAskedItem(rs.Asked),
		}
		for _, l := range rs.Landed {
			run.landed[l.Seat] = copyUUIDs(l.Cards)
		}
		out[rs.ID] = run
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// unknownPromptRunKeys lists the runs on disk whose continuation this
// binary does not have.
func unknownPromptRunKeys(in []promptRunSnapshot) []string {
	var out []string
	for _, rs := range in {
		if !KnownPromptRunThen(rs.Key) {
			out = append(out, "prompted-run continuation "+rs.Key)
		}
	}
	return out
}
