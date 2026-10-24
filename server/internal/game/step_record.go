package game

import (
	"fmt"
	"sort"

	"github.com/google/uuid"
)

// step_record.go — what an earlier step hands a later one, as data
// (#4163, CR 608.2c).
//
// # The gap
//
// A step that waits for an earlier step's prompt may run after a
// restore, so it cannot read anything the earlier step left in a Go
// variable. Most later steps re-read what they need from the game (the
// cards drawn this turn, the life total). Some need something the game
// keeps no record of: the number of cards a player discarded (Windfall
// draws that many), the Soldiers a spell just made (Martial Coup spares
// them), the number a player chose for a continuation (Arcane Denial).
//
// # The record
//
// A step writes a small record, keyed by the resolving item and a name
// the card chooses, and a later step of the same resolution reads it:
// a list of object or player IDs, a list of numbers and a list of words
// (#4228). It is plain
// data, carried by Clone and the snapshot (GameSnapshot.StepRecords,
// additive within v9), and it is no closure, so the closure ratchet
// does not move.
//
// It lives for one resolution: every record is dropped as the
// resolution closes (holdForOpenResolutionLocked) and as a new one
// opens, so a record is never read by another item. A continuation that
// runs after its resolution has closed (a prompt that does not block the
// table, answered later) writes and reads in one call and cannot park,
// so it never needs one across a restore.

// StepRecord is what one step hands a later one.
//
// Strs is a list of words (#4228): the counter kinds Rhys, the Evermore
// still has to ask about. A new field of the record, so additive within
// the snapshot version: `omitempty`, and a binary from before it reads
// a record without it exactly as one that never had any. Nothing about
// it is a closure.
type StepRecord struct {
	IDs  []uuid.UUID `json:"ids,omitempty"`
	Ns   []int       `json:"ns,omitempty"`
	Strs []string    `json:"strs,omitempty"`
}

// N is the record's first number, 0 when it has none.
func (r StepRecord) N() int {
	if len(r.Ns) == 0 {
		return 0
	}
	return r.Ns[0]
}

// stepRecordEntry is one record, with the item and the name it is
// filed under. It is its own snapshot form.
type stepRecordEntry struct {
	Item   uuid.UUID  `json:"item"`
	Key    string     `json:"key"`
	Record StepRecord `json:"record"`
}

func cloneStepRecord(r StepRecord) StepRecord {
	return StepRecord{IDs: copyUUIDs(r.IDs), Ns: copyInts(r.Ns), Strs: copyStrings(r.Strs)}
}

func cloneStepRecords(in []stepRecordEntry) []stepRecordEntry {
	if len(in) == 0 {
		return nil
	}
	out := make([]stepRecordEntry, len(in))
	for i, e := range in {
		out[i] = stepRecordEntry{Item: e.Item, Key: e.Key, Record: cloneStepRecord(e.Record)}
	}
	return out
}

func stepRecordItemID(item *StackItem) uuid.UUID {
	if item == nil {
		return uuid.Nil
	}
	return item.ID
}

// RecordStepForEffect files rec under item and key for a later step of
// the same resolution, replacing any record already there. The slice is
// replaced, never written in place, so a clone that shares it keeps its
// own.
//
// Caller holds g.mu (write).
func (g *Game) RecordStepForEffect(item *StackItem, key string, rec StepRecord) {
	id := stepRecordItemID(item)
	out := make([]stepRecordEntry, 0, len(g.stepRecords)+1)
	for _, e := range g.stepRecords {
		if e.Item != id || e.Key != key {
			out = append(out, e)
		}
	}
	g.stepRecords = append(out, stepRecordEntry{Item: id, Key: key, Record: cloneStepRecord(rec)})
}

// StepRecordForEffect is the record filed under item and key, and
// whether there is one. The result is a copy.
//
// Caller holds g.mu.
func (g *Game) StepRecordForEffect(item *StackItem, key string) (StepRecord, bool) {
	id := stepRecordItemID(item)
	for _, e := range g.stepRecords {
		if e.Item == id && e.Key == key {
			return cloneStepRecord(e.Record), true
		}
	}
	return StepRecord{}, false
}

// dropStepRecordsLocked forgets every record: a resolution has closed
// or a new one has opened.
func (g *Game) dropStepRecordsLocked() {
	g.stepRecords = nil
}

// thenStepKeys are the continuation keys registered as steps (#4162).
// Guarded by effectRegistryMu, with the other effect keys, because the
// keys share one namespace and one ledger ("steps <key>").
var thenStepKeys = map[string]bool{}

// RegisterThenStepsKey registers key as a continuation written as
// steps. The effects package calls it from effects.RegisterThenSteps,
// which holds the steps. Panics on a bad or duplicate key. The key is
// an on-disk identity: never renamed, never reused.
func RegisterThenStepsKey(key string) {
	effectRegistryMu.Lock()
	defer effectRegistryMu.Unlock()
	checkEffectKey("continuation steps", key)
	if thenStepKeys[key] {
		panic(fmt.Sprintf("game: continuation steps %q are already registered", key))
	}
	thenStepKeys[key] = true
}

// KnownThenSteps reports whether this binary has continuation steps
// registered under key.
func KnownThenSteps(key string) bool {
	if key == "" {
		return false
	}
	effectRegistryMu.RLock()
	defer effectRegistryMu.RUnlock()
	return thenStepKeys[key]
}

// registeredThenStepKeysLocked lists the ledger lines, "steps <key>".
// Caller holds effectRegistryMu.
func registeredThenStepKeysLocked() []string {
	out := make([]string, 0, len(thenStepKeys))
	for k := range thenStepKeys {
		out = append(out, "steps "+k)
	}
	sort.Strings(out)
	return out
}
