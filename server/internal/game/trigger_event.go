package game

import "github.com/google/uuid"

// trigger_event.go — #1223: the triggering event, carried as DATA on
// the triggered ability's stack item (CR 603.2, CR 603.10).
//
// # The gap
//
// A `TriggeredAbility` is handed the event twice — once in
// `AppliesTo`, once in `Build` — and both hands go away immediately.
// `AppliesTo` returns a bool. `Build` returns a `*StackItem`, and the
// only way to get the event past it is to CAPTURE it in the `Effect`
// closure — which is the catalog's established convention, and a
// perfectly good place for a value right up until somebody asks one
// of these two questions. This seam is both of them:
//
//  1. THE TARGET CLAUSE. A trigger's targets are chosen as it is put
//     on the stack (CR 603.3d), by the harvester, from
//     `TriggeredAbility.Targets` — which is STATIC catalog data
//     declared before the game started. Scrap Trawler's "target
//     artifact card in your graveyard with lesser mana value" cannot
//     be stated at all, because "lesser" is lesser than the mana
//     value of the artifact that just died and no closure runs
//     between the event and the clause.
//
//  2. THE COPY. CR 707.10 copies a triggered ability together with
//     "any choices made when it triggered", and the ruling for
//     Strionic Resonator is explicit that the copy remembers the same
//     triggering event. An `Effect` closure reached the copy by
//     accident of sharing the func pointer; nothing said so, and
//     nothing could say so for a clause, a prompt's resume frame or a
//     restored snapshot.
//
// # The shape
//
// ONE typed value, `TriggerContext`, stamped onto `StackItem.Trigger`
// at the moment the item is built, read by the effect through
// `effects.Context.Trigger()` and by the clause through
// `TriggeredAbility.TargetsFrom`. It is plain DATA — no closures, no
// pointers into zone slices — for the three reasons the rest of the
// item's announce-time record is data (`Paid`, `XValue`, `Modes`):
//
//   - it survives a COPY, because copying it is copying a struct;
//   - it survives a SNAPSHOT, so a game paused on an unanswered
//     trigger prompt is still a restore point;
//   - it is a LOOKUP rather than a scan. The catalog's other way of
//     reaching a past event is to walk `g.Events` backwards for the
//     last one of a kind, which finds the wrong event the moment two
//     of that kind land in one batch (two creatures dying to one
//     wrath). No card does that for its OWN triggering event today,
//     and this is what keeps it that way.
//
// # Last-known information (CR 603.10)
//
// Half of what a trigger wants to know is about an object that has
// already gone: the creature that died, the artifact that was put
// into a graveyard, the permanent that left. CR 603.10 says such an
// ability is judged on what the object looked like immediately before
// the event. The engine already keeps that snapshot — the harvester's
// `lastKnownBattlefield` map, written by every battlefield-exit site
// and DELETED as soon as the LTB harvest finishes — so `Object` is
// read off it while it is still there and copied by value. A card
// that read it later would find a hole.

// ObjectSnapshot is last-known information about ONE object the
// triggering event was about: the creature that died, the permanent
// that entered, the card that was discarded.
//
// A flat value rather than a `Card` or a `*Card`, and that is the
// whole design:
//
//   - a POINTER into a zone slice is invalid the moment anything
//     moves, which for an LTB trigger is already true when the
//     trigger is built;
//   - a whole `Card` carries `effective`, `PrintedSelf` and the rest
//     of the layer bookkeeping, none of which means anything once the
//     object is gone, and all of which would have to be snapshotted
//     to keep the stack item restorable.
//
// What is here is what a printed clause asks about an object that has
// left: its name, its identity, its types, its mana value and whose
// it was. A clause that wants something else asks for the field to be
// added rather than reaching for a pointer — see the note on power
// and toughness below for what that costs when the answer is not
// already in the snapshot.
type ObjectSnapshot struct {
	// ID is the object's instance id. Still meaningful after the
	// object has gone — a card in a graveyard keeps it — which is
	// what lets "return it to its owner's hand" find the thing.
	ID uuid.UUID

	// OracleID is the catalog identity, for a clause that compares
	// card identity rather than characteristics.
	OracleID string

	// Name is the object's name as it was (CR 603.10), which is not
	// necessarily the name printed on the card underneath — a copy
	// effect or a face-down permanent both make them differ.
	Name string

	// Types / Subtypes / Supertypes are the POST-LAYER type line as
	// it was — the same lists `Characteristic` carries, with the type
	// line's own capitalisation ("Creature"). Read them through
	// HasType or PermanentTypes rather than comparing directly; both
	// answer in the lowercase tokens the rest of the engine uses.
	Types      []string
	Subtypes   []string
	Supertypes []string

	// Colors are the object's colours as they were, WUBRG-ordered.
	Colors []string

	// ManaValue is CR 202.3's mana value of the card the object was.
	// Scrap Trawler's "with lesser mana value" is one comparison
	// against this, and it is the number that is unrecoverable a
	// moment later: the card is in a graveyard, where a face-down or
	// copied permanent's mana value is not the one it had.
	ManaValue int

	// DELIBERATELY ABSENT: power and toughness. "A creature with
	// power equal to the power of the creature that died" is a real
	// printed clause and it is not answerable here yet, because the
	// only last-known-information store the engine keeps
	// (Game.lastKnownBattlefield) holds a Characteristic, and the
	// engine applies +1/+1 and -1/-1 counters OUTSIDE it
	// (Card.PowerForComparison) — while MoveCard clears a permanent's
	// counters on the way out of the battlefield (zone.go), so by the
	// time a dies-trigger is harvested they are gone from both
	// places. A Power field here would silently read 2 for the 3/3 a
	// counter made, which is the one case such a clause is written
	// about. It waits for a second LKI store or for counters to move
	// into the snapshot, and neither is this seam's to decide — see
	// docs/decisions/0072-protection.md §2 on the cost of a second
	// store.

	// Controller is who controlled the object, and Owner who owned
	// it. Both as they were: control can change on the way out and
	// the rules read the pre-event value.
	Controller, Owner uuid.UUID

	// Token records that the object was a token (CR 111). A clause
	// that says "nontoken permanent" — Cloudstone Curio prints one —
	// reads it here, because a token that has left the battlefield
	// has ceased to exist and there is nothing to ask.
	Token bool
}

// HasType reports whether the object had the lowercase card type
// `t` ("creature", "artifact", "land").
func (o *ObjectSnapshot) HasType(t string) bool {
	if o == nil {
		return false
	}
	return typeListHas(o.Types, t)
}

// PermanentTypes is the object's card types narrowed to the ones that
// are PERMANENT types (CR 110.4a: artifact, battle, creature,
// enchantment, land, planeswalker), in the order the type line had
// them.
//
// "Shares a permanent type with it" (Cloudstone Curio) is an
// intersection of two of these lists, and it has to be this list
// rather than Types: an artifact creature that entered shares a type
// with a plain artifact, while an INSTANT card in a graveyard shares
// nothing with anything on the battlefield and must not be admitted
// by a loop that forgot to filter.
//
// Returns the CANONICAL LOWERCASE token ("creature"), not the type
// line's own capitalisation, because every type predicate in the
// engine — Card.HasCardType and the catalog's Creature() / Artifact()
// — takes a lowercase needle and compares it against a lowercased
// type line. Handing back "Creature" would silently match nothing.
func (o *ObjectSnapshot) PermanentTypes() []string {
	if o == nil {
		return nil
	}
	var out []string
	for _, t := range o.Types {
		if canon, ok := permanentTypeOf(t); ok {
			out = append(out, canon)
		}
	}
	return out
}

// SharesPermanentTypeWith reports whether the object and `c` have a
// permanent type in common (CR 110.4a).
func (o *ObjectSnapshot) SharesPermanentTypeWith(c Card) bool {
	for _, t := range o.PermanentTypes() {
		if c.HasCardType(t) {
			return true
		}
	}
	return false
}

// permanentTypeOf maps a card type to its canonical lowercase token
// when it is one of CR 110.4a's six permanent types.
//
// Kindred / tribal is deliberately absent: a Kindred card is always
// also one of the six, so a card that shared ONLY "kindred" with
// something does not exist, and admitting it would let an
// enchantment share a type with an artifact.
//
// Case-insensitive, because a Characteristic's Types carry the type
// line's own capitalisation ("Creature") while every predicate in the
// engine compares against a lowercase token — the same asymmetry
// Card.HasCardType resolves with typeListHas.
func permanentTypeOf(t string) (string, bool) {
	for _, p := range permanentTypes {
		if equalFoldASCII(t, p) {
			return p, true
		}
	}
	return "", false
}

// permanentTypes is CR 110.4a's list, in the order a type line prints
// them.
var permanentTypes = []string{"artifact", "battle", "creature", "enchantment", "land", "planeswalker"}

// TriggerContext is what a triggered ability knows about the event
// that fired it: the event itself, and last-known information about
// the object it was about.
//
// Handed to the ability's target clause (`TriggeredAbility.TargetsFrom`)
// as the trigger is put on the stack, and to its effect
// (`effects.Context.Trigger()`) when the item resolves. The zero value
// is the honest answer for every item that is not a harvested trigger
// — a cast spell, an activated ability, a reflexive trigger, and a
// trigger restored from a snapshot written before this field existed
// — and `Fired` is how a card asks.
type TriggerContext struct {
	// Event is the triggering event, verbatim and by value. Every
	// field of it is a fact about what happened: `Amount` is the
	// damage dealt or the life gained or the counters placed,
	// `CardID` names the object, `StackItemID` names the spell or
	// ability, `OldZone` / `NewZone` say where it went, `Combat`
	// says whether the damage was combat damage.
	//
	// The whole event rather than a hand-picked subset, because the
	// alternative is a new field on this struct for every clause
	// that wants one more fact, and because `Event` is already the
	// flat, closure-free, JSON-tagged shape the log is built out of.
	Event Event

	// Object is CR 603.10 last-known information about the object the
	// event was about — `Event.CardID`, which for a zone change is
	// the card that moved and for an ETB is the permanent that
	// entered. Nil when the event names no object, and when the
	// object could not be found in any zone.
	Object *ObjectSnapshot
}

// Fired reports whether this context describes a real triggering
// event. False for the zero value, which is what every non-trigger
// item carries.
func (tc TriggerContext) Fired() bool { return tc.Event.Kind != "" }

// Amount is the event's count payload — the damage dealt, the life
// gained or lost, the cards drawn, the counters placed. Zero for an
// event that has none, and for a context that never fired.
//
// The named read for the "whenever ~ deals damage to a player, …
// that much" family, which is the printed shape this seam was filed
// for.
func (tc TriggerContext) Amount() int { return tc.Event.Amount }

// triggerContextLocked builds the context for one harvested trigger.
//
// Called ONCE per trigger instance, at the moment the harvester
// decides the ability applies, and carried from there through every
// prompt frame to the built item. Not re-derived later, and that is
// the point: `lastKnownBattlefield` is deleted as soon as the LTB
// harvest for an event finishes (`harvestLTB`'s deferred delete), so
// a second call a prompt later would produce a DIFFERENT and emptier
// answer for exactly the triggers that need one.
//
// Caller must hold g.mu.
func (g *Game) triggerContextLocked(ev Event) TriggerContext {
	return TriggerContext{Event: ev, Object: g.objectSnapshotLocked(ev.CardID)}
}

// objectSnapshotLocked reads last-known information about one object
// (CR 603.10), preferring the battlefield-exit snapshot over the
// live card.
//
// The ORDER is the rule. `lastKnownBattlefield` holds what the
// permanent looked like immediately before it left, which is what a
// dies-trigger is judged on; the live card is what it looks like now,
// in a graveyard, with its layer cache dropped and its continuous
// effects gone. For an ETB or any other event about an object that is
// still where it was, there is no snapshot and the live read is the
// same answer.
//
// The two halves come from different places on purpose. Types,
// colours, name and P/T are CHARACTERISTICS and come off the
// snapshot; the oracle id, the owner and the mana value are facts
// about the CARD and come off the card, because the mana value of a
// permanent is computed from its printed cost (CR 202.3b) and
// `Characteristic` does not carry one.
//
// Returns nil when the event names no object or the object is in no
// zone the engine can see (a token that has ceased to exist).
//
// Caller must hold g.mu.
func (g *Game) objectSnapshotLocked(cardID uuid.UUID) *ObjectSnapshot {
	if cardID == uuid.Nil {
		return nil
	}
	card := g.findCardByIDLocked(cardID)
	if card == nil {
		return nil
	}
	out := &ObjectSnapshot{
		ID:         cardID,
		OracleID:   card.OracleID,
		Name:       card.Name,
		ManaValue:  card.ManaValue(),
		Owner:      card.Owner,
		Controller: card.Controller,
		Token:      card.IsToken(),
	}
	ch, ok := g.lastKnownBattlefield[cardID]
	if !ok {
		ch = card.Effective()
	}
	out.Types = copyStrings(ch.Types)
	out.Subtypes = copyStrings(ch.Subtypes)
	out.Supertypes = copyStrings(ch.Supertypes)
	out.Colors = copyStrings(ch.Colors)
	if ch.Name != "" {
		out.Name = ch.Name
	}
	if ch.Controller != uuid.Nil {
		out.Controller = ch.Controller
	}
	return out
}

// cloneTriggerContext deep-copies a context for the undo clone and
// the ability copy. The slices get their own backing arrays for the
// reason `StackItem.Targets` does: a restore that aliased them would
// let one game mutate the other.
func cloneTriggerContext(tc *TriggerContext) *TriggerContext {
	if tc == nil {
		return nil
	}
	out := *tc
	out.Event.Colors = copyStrings(tc.Event.Colors)
	if tc.Object != nil {
		obj := *tc.Object
		obj.Types = copyStrings(tc.Object.Types)
		obj.Subtypes = copyStrings(tc.Object.Subtypes)
		obj.Supertypes = copyStrings(tc.Object.Supertypes)
		obj.Colors = copyStrings(tc.Object.Colors)
		out.Object = &obj
	}
	return &out
}

// triggerTargetsLocked is the clause list an ability is announced
// under — `TargetsFrom` when the card declares one, the static
// `Targets` otherwise.
//
// Called at every point the dispatch reads the ability's clause, and
// called AGAIN rather than cached in a frame, deliberately: the
// answer is a function of the trigger context (which is carried and
// therefore stable) and of the board (which is not, and must not be
// — a clause re-read after an optional prompt should see the board as
// it is, exactly as `refreshTargetChoicesLocked` re-reads the legal
// set for the same reason).
//
// Caller must hold g.mu.
func (g *Game) triggerTargetsLocked(t TriggeredAbility, tc TriggerContext, source *Card) *TargetSpec {
	if t.TargetsFrom != nil {
		return t.TargetsFrom(tc, source, g)
	}
	return t.Targets
}
