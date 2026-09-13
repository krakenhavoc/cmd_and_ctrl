package game

import "github.com/google/uuid"

// creature_type_choice.go — "as this permanent enters, choose a
// creature type" (CR 614.12): Cavern of Souls, Door of Destinies,
// Vanquisher's Banner, Adaptive Automaton.
//
// The prompt is queued from the permanent's ETB hook rather than from
// the CR 614 replacement pipeline, and that is a declared
// simplification worth stating precisely, because the file next door
// (entry_choice.go) proves the pipeline CAN pause for a question.
//
// It can pause only for a LAND. `ReplacementEvent.entryResumable` is
// set on exactly one entry site — the land-play branch — because the
// generic resume would finish a spell's entry without the rest of its
// resolution. Three of the four cards above are artifacts, so a
// pipeline-based choice would have to work one way for Cavern of
// Souls and another way for Door of Destinies, which is two
// mechanisms for one printed phrase.
//
// What the simplification costs: the permanent is on the battlefield
// for the window between entering and the answer arriving, with an
// empty NamedTribe. Nothing can act in that window — an outstanding
// PendingChoice stops priority from passing, which is the same
// property every other prompt in the engine relies on — so the
// difference is unobservable short of a second permanent entering
// simultaneously, which the engine does not do. The direction is also
// the safe one: a static ability reading an empty NamedTribe applies
// to nothing rather than to everything.

// PendingChoiceCreatureType is the "choose a creature type" prompt.
// Answered with a `{creature_type: "Elf"}` payload; the chooser is the
// entering permanent's controller.
//
// Unlike every other Options-bearing kind, the legal set is not stored
// on the choice: it is the whole CR 205.3m vocabulary, identical for
// every such prompt, and the view materialises it from
// AllCreatureTypes at serialisation time. Storing 345 strings per
// prompt would put them through clone, the snapshot encoder and the
// drift test for no information the server does not already have.
const PendingChoiceCreatureType PendingChoiceKind = "choose_creature_type"

// EventCreatureTypeChosen records a CR 614.12 answer in the event
// log. `Label` carries the chosen type and `CardID` the permanent it
// was chosen for, so the log reads "Cavern of Souls — Elf".
const EventCreatureTypeChosen EventKind = "creature_type_chosen"

// QueueCreatureTypeChoiceForEffect queues the "choose a creature
// type" prompt for `chooser`, on behalf of the permanent `source`.
// Returns the choice ID.
//
// Caller must hold g.mu (an ETB hook does).
func (g *Game) QueueCreatureTypeChoiceForEffect(chooser, source uuid.UUID, reason string) uuid.UUID {
	return g.QueueChoiceForEffect(PendingChoice{
		Kind:    PendingChoiceCreatureType,
		Chooser: chooser,
		Count:   1,
		Source:  source,
		Reason:  reason,
	})
}

// ResolveCreatureTypeChoice processes a resolve_choice action for a
// PendingChoiceCreatureType entry: the chooser names a creature type
// and it is stamped onto the source permanent's NamedTribe.
//
// The type is validated against the CR 205.3m vocabulary and
// normalised to its canonical spelling, so a client that sends "elf"
// or an unknown string cannot put junk into game state that every
// later recompute then compares against.
//
// The permanent is located live rather than trusted from the queue:
// the prompt is asynchronous and the permanent can have left in the
// meantime (a Cavern of Souls destroyed in response to its own entry
// is a legal, if odd, board). A missing source is NOT an error — the
// choice is still made and simply has nowhere to land, which is what
// CR 608.2 does with an effect whose object has gone.
//
// Caller must NOT hold g.mu.
func (g *Game) ResolveCreatureTypeChoice(choiceID, chooserID uuid.UUID, creatureType string) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	idx := -1
	for i, c := range g.PendingChoices {
		if c != nil && c.ID == choiceID {
			idx = i
			break
		}
	}
	if idx < 0 {
		return ErrPendingChoiceNotFound
	}
	choice := g.PendingChoices[idx]
	if choice.Kind != PendingChoiceCreatureType {
		return ErrInvalidParam
	}
	if choice.Chooser != chooserID {
		return ErrNotTheChooser
	}
	canonical, ok := CanonicalCreatureType(creatureType)
	if !ok {
		return ErrInvalidParam
	}
	g.dequeueChoiceLocked(idx)

	if i := findCardOnBattlefield(g, choice.Source); i >= 0 {
		g.Battlefield.Cards[i].NamedTribe = canonical
		// The named tribe is an AppliesTo input for every static
		// ability on the source (Door of Destinies' anthem, Adaptive
		// Automaton's own type-add), and the layer engine caches its
		// resolution until something invalidates it. Nothing else in
		// this path emits an event the layer listener watches, so the
		// bump is explicit — without it the lord's +1/+1 would appear
		// only when some unrelated permanent happened to move.
		g.layerVersion.Add(1)
	}
	g.EmitEvent(Event{
		Kind:   EventCreatureTypeChosen,
		Actor:  chooserID,
		CardID: choice.Source,
		Label:  canonical,
	})
	g.runStateChecksLocked()
	return nil
}

// NamedTribeOf returns the creature type chosen for the permanent
// `sourceID` currently on the battlefield, or "" when none has been
// chosen (or the permanent is gone).
//
// The accessor exists so a catalog card's AppliesTo predicate reads
// the tribe off the SOURCE card it was handed rather than re-finding
// it, and so the one place that answers "what did they name" is
// shared by the four cards that ask.
//
// Caller must hold either lock.
func (g *Game) NamedTribeOf(sourceID uuid.UUID) string {
	if i := findCardOnBattlefield(g, sourceID); i >= 0 {
		return g.Battlefield.Cards[i].NamedTribe
	}
	return ""
}
