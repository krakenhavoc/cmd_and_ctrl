package game

import (
	"sort"

	"github.com/google/uuid"
)

// stack_lki.go — last-known information for a spell that has LEFT the
// stack without resolving (#1255, CR 608.2h, CR 707.10).
//
// # The rule
//
// CR 608.2h: when an effect refers to an object that has left the zone
// it was expected in, it uses the object's last known information. A
// storm trigger says "copy it for each other spell cast before it this
// turn" — "it" is the storm spell, and the trigger exists independently
// of that spell (CR 113.7a). So when the spell is countered in response
// to its own trigger, the copies are still made, from the spell as it
// last existed on the stack. The same is true of every copy effect that
// NAMES the spell without targeting it: Thousand-Year Storm, Doublecast's
// "copy that spell", Breeches's "copy that spell".
//
// # Why the lookup is opt-in
//
// A copy effect that TARGETS the spell — Reverberate, Twincast, Dualcaster
// Mage, Lithoform Engine — is governed by CR 608.2b instead: when its
// target has left the stack the target is illegal, and a spell or
// ability whose every target is illegal does nothing at all. The engine
// already gets that right (the CR 608.2b re-check counters the item
// before its effect runs), and CopySpellForEffect's ErrCardNotFound is
// the backstop for a multi-target item. Widening that lookup would
// resurrect a target the rules say is gone. So the strict entry point is
// unchanged and CopyLastKnownSpellForEffect is a second one, called only
// by effects that do not target the spell.
//
// # The record
//
// Game.lastKnownStack holds a VALUE copy of the card and its stack item,
// taken at the one choke point every non-resolving exit from the stack
// passes through — routeCardToZoneLocked with a card whose source zone
// is the stack, which is counterspells, Remand-style returns, "exile
// target spell", airbend, and the sandbox's manual move (#1318 made it
// the source zone rather than a DropStackMeta flag only some exits set). A
// spell that RESOLVES does not need an entry: the only effect that can
// copy it after that point is its own ("copy this spell"), and #920's
// Game.resolving slot already answers for it.
//
// It lives for the rest of the turn and is cleared at the turn boundary
// with the rest of the per-turn state. That is long enough by
// construction: a copy effect that names a spell is a stack object
// created while the spell was on the stack, and the stack must be empty
// before the turn can end. It is bounded by the turn's cast count.
//
// Clone carries it (an undo across a counterspell rewinds it with the
// spell), and since ADR 0041 phase 3 tier 4 (#1497, P9) so does the
// snapshot. Before it every reader of the record was a stack item or
// delayed trigger whose behaviour is a closure — so a snapshot that
// could need the record was not a restore point anyway — and the
// snapshot dropped it. Tier 4 makes those readers data, and a restore
// that dropped the record would make a restored storm trigger miss the
// copies of a countered spell. The card is carried like any card and
// the item like a spell item (GameSnapshot.LastKnownStack).

// lastKnownSpell is one spell as it last existed on the stack.
type lastKnownSpell struct {
	card Card
	item StackItem
}

// rememberLeavingSpellLocked records the spell `cardID` as it stands on
// the stack, just before it leaves without resolving. A card that is not
// on the stack, or has no stack item, records nothing. Caller must hold
// g.mu.
func (g *Game) rememberLeavingSpellLocked(cardID uuid.UUID) {
	if g.Stack == nil {
		return
	}
	item, ok := g.StackMeta[cardID]
	if !ok || item == nil || item.Kind != StackItemSpell {
		return
	}
	for _, c := range g.Stack.Cards {
		if c.InstanceID != cardID {
			continue
		}
		if g.lastKnownStack == nil {
			g.lastKnownStack = make(map[uuid.UUID]lastKnownSpell)
		}
		g.lastKnownStack[cardID] = lastKnownSpell{card: c, item: *item}
		return
	}
}

// lastKnownSpellLocked returns the last-known card and a FRESH copy of
// the stack item of a spell that has left the stack, and false when the
// turn has no record of it. The item is a copy so that nothing the copy
// path does to it can reach the record. Caller must hold g.mu.
func (g *Game) lastKnownSpellLocked(spellID uuid.UUID) (Card, *StackItem, bool) {
	rec, ok := g.lastKnownStack[spellID]
	if !ok {
		return Card{}, nil, false
	}
	item := rec.item
	return rec.card, &item, true
}

// clearLastKnownStackLocked forgets the turn's records. Called at the
// turn boundary. Caller must hold g.mu.
func (g *Game) clearLastKnownStackLocked() {
	g.lastKnownStack = nil
}

// cloneLastKnownStack copies the record for Clone. The values are never
// mutated after they are stored, so a per-entry value copy suffices —
// the same argument lastKnownBattlefield's clone makes.
func cloneLastKnownStack(src map[uuid.UUID]lastKnownSpell) map[uuid.UUID]lastKnownSpell {
	if len(src) == 0 {
		return nil
	}
	out := make(map[uuid.UUID]lastKnownSpell, len(src))
	for k, v := range src {
		out[k] = v
	}
	return out
}

// CopyLastKnownSpellForEffect is CopySpellForEffect for a copy effect
// that does NOT target the spell it copies (CR 608.2h): the spell is
// copied from the stack when it is still there, and otherwise from its
// last-known information — the storm spell countered in response to its
// own trigger, the spell Doublecast's trigger names countered before
// the trigger resolves.
//
// Everything past the lookup is CopySpellForEffect's: the "except"
// clause, the CR 707.10c re-target offer and the copy itself already
// work entirely from VALUE copies of the card and its item, which is
// what lets the Chain cycle copy from a graveyard (#920) and lets this
// copy from a counterspell's victim.
//
// A copy effect that targets the spell must call CopySpellForEffect,
// never this: see the file comment.
//
// Errors: ErrCardNotFound when the spell is neither on the stack, nor
// resolving, nor recorded this turn. Caller must hold g.mu.
func (g *Game) CopyLastKnownSpellForEffect(spellID, controller uuid.UUID, mayChooseNewTargets bool, except func(v *PrintedValues)) error {
	src, item, ok := g.stackSpellLocked(spellID)
	if !ok {
		src, item, ok = g.lastKnownSpellLocked(spellID)
	}
	if !ok {
		return ErrCardNotFound
	}
	g.copySpellFromLocked(src, item, controller, mayChooseNewTargets, except)
	return nil
}

// lastKnownSpellSnapshot is one Game.lastKnownStack entry on disk: the
// card and its stack item, both through the ordinary mirrors.
type lastKnownSpellSnapshot struct {
	Card cardSnapshot      `json:"card"`
	Item stackItemSnapshot `json:"item"`
}

// snapshotLastKnownStack is the capture half, sorted by card ID so two
// captures of one state are byte-identical. The census sees each entry
// the way it sees a live stack item and card: a spell has no closure of
// its own, so an entry blocks nothing unless it holds something no
// restore could rebuild.
func snapshotLastKnownStack(m map[uuid.UUID]lastKnownSpell, cen *ContinuationCensus) []lastKnownSpellSnapshot {
	if len(m) == 0 {
		return nil
	}
	ids := make([]uuid.UUID, 0, len(m))
	for id := range m {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i].String() < ids[j].String() })
	out := make([]lastKnownSpellSnapshot, 0, len(ids))
	for _, id := range ids {
		rec := m[id]
		item := rec.item
		out = append(out, lastKnownSpellSnapshot{
			Card: snapshotCard(rec.card, cen),
			Item: snapshotStackItemAs(&item, rec.card.OracleID, cen),
		})
	}
	return out
}

// restoreLastKnownStack is the restore half. The key is the card's
// instance ID, exactly as rememberLeavingSpellLocked files it.
func restoreLastKnownStack(in []lastKnownSpellSnapshot) map[uuid.UUID]lastKnownSpell {
	if len(in) == 0 {
		return nil
	}
	out := make(map[uuid.UUID]lastKnownSpell, len(in))
	for i := range in {
		card := restoreCard(&in[i].Card)
		// A spell is never a stamped ability, so the Q3 flag the
		// second result carries cannot be set here.
		item, _ := restoreStackItem(&in[i].Item)
		out[card.InstanceID] = lastKnownSpell{card: card, item: *item}
	}
	return out
}
