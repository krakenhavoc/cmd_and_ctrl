package game

import (
	"github.com/google/uuid"
)

// mana_trigger.go — TRIGGERED MANA ABILITIES (CR 605.1b), the one kind
// of trigger that never reaches the stack. See ADR 0074.
//
// CR 605.1b makes a triggered ability a mana ability when three things
// are true at once: it triggers off an activated mana ability
// resolving, it does not target, and it could add mana. CR 605.4a then
// says a mana ability does not use the stack — it simply resolves,
// with no priority window for anybody. "Whenever enchanted land is
// tapped for mana, its controller adds an additional {G}" is the
// printed shape; Wild Growth, Overgrowth, Utopia Sprawl, Fertile
// Ground, Mana Flare, Mirari's Wake and Crypt Ghast are the family.
//
// # Why this is not a TriggeredAbility with a flag
//
// Nothing on TriggeredAbility survives the change. Watches, AppliesTo's
// Event, Build's *StackItem, OptionalPrompt, Targets, HasLegalTarget,
// OncePerBatch, BatchKey, FromStack and Zones all exist to get an item
// onto the stack correctly, and a mana trigger has no item, no targets
// (CR 605.1b forbids them), no prompt and no batch. A flagged
// TriggeredAbility would be nine fields the harvester has to remember
// to skip, one `if` away from a Wild Growth taking a priority window
// CR 605.4a does not allow.
//
// What it keeps is the two things that are not about the stack: the
// CR 716/719/721 designation gate (ActiveWhen, evaluated in
// ManaTriggersForCard and nowhere else) and CatalogAbilityKey, so a
// CR 613.1f ability-removing effect on the Aura takes the trigger away.
//
// # What fires it
//
// One function — fireManaTriggersLocked — called from exactly three
// places, all of them a permanent being TAPPED for mana:
//
//   - ActivateManaAbility, after the rider, when the activation put at
//     least one token in a pool directly;
//   - ResolveManaChoice, when the answered pick was a mana ability's
//     (PendingChoice.ManaTapped);
//   - materializePlanLocked, once per planned source.
//
// A mana ability either mints directly or queues a pick; no printed
// card does both in one ability, so each real card takes exactly one
// of the first two branches and fires exactly once. AddMana from a
// resolving spell (Dark Ritual) is not "tapped for mana" (CR 106.12a)
// and never fires: AddManaForEffect does not call this, and the picks
// it queues leave ManaTapped false.
//
// Triggered mana does not re-trigger: addTriggeredManaLocked never
// calls the firing function. That is also the rule — nothing was
// tapped for the extra {G}, so a second Wild Growth does not see it.

// ManaProduced is one completed "tapped for mana" production (CR
// 106.12a): a permanent's mana ability resolved and its mana landed in
// a pool. The payload every ManaTrigger is asked about.
type ManaProduced struct {
	// Source is the permanent that was tapped, as it was at the moment
	// it produced. A VALUE, not a pointer: a tap-cost ability may also
	// sacrifice its source (Lotus Petal), and the predicate still has
	// to be able to ask what it was.
	Source Card

	// Controller is the player whose pool the mana landed in, and the
	// player a triggered mana ability adds its mana to. Every printed
	// card agrees on that — "its controller adds", "that player adds",
	// and Mirari's Wake's bare "add", which reaches the same player
	// through its own AppliesTo.
	Controller uuid.UUID

	// Colors are the mana types added, in the order added, as the
	// single uppercase symbols ManaToken.Color uses ("W", "U", "B",
	// "R", "G", "C"). This is the produced colour "one mana of any
	// type that land produced" reads (Mana Flare, Mirari's Wake,
	// Zendikar Resurgent) and the colour Forsaken Monument's "if that
	// mana is {C}" tests.
	Colors []string
}

// SourceID is the tapped permanent's instance ID.
func (p ManaProduced) SourceID() uuid.UUID { return p.Source.InstanceID }

// ProducedColorPipe renders p.Colors as a produced-mana string
// offering one mana of any type produced — "{U}" for a single type,
// "{U|G}" for two. The whole of "add one mana of any type that land
// produced"; an empty production renders "", which adds nothing.
//
// Duplicates collapse, because two {G} tokens do not make green a
// wider choice.
func (p ManaProduced) ProducedColorPipe() string {
	seen := map[string]bool{}
	out := ""
	for _, c := range p.Colors {
		if c == "" || seen[c] {
			continue
		}
		seen[c] = true
		if out != "" {
			out += "|"
		}
		out += c
	}
	if out == "" {
		return ""
	}
	return "{" + out + "}"
}

// ManaTrigger is a TRIGGERED MANA ABILITY (CR 605.1b): it triggers off
// a permanent being tapped for mana and it adds mana, so it resolves
// immediately and never uses the stack (CR 605.4a).
//
// It lives on Spec.ManaTriggers, NOT Spec.Triggered. The test is one
// line: if the ability adds mana and does not target, it belongs here.
// Electro's and Fire Nation Palace's "add mana" triggers fire on a
// cast and on an attack rather than on a mana ability, so they are
// ordinary stack triggers (CR 605.5a) and stay in Spec.Triggered.
//
// Both callbacks run under g.mu held in write mode, from the mana
// production path. READ-ONLY: use *ForEffect accessors, never a public
// locking mutator.
type ManaTrigger struct {
	// Label is what the log and the client call this ability.
	Label string

	// AppliesTo decides whether this production triggers this source.
	// `source` is the permanent carrying the trigger (the Aura, the
	// Mana Flare), live on the battlefield. Nil means "every
	// production", which no printed card wants.
	AppliesTo func(prod ManaProduced, source *Card, g *Game) bool

	// Produced is the mana this trigger adds, in the same
	// ParseProducedMana grammar a mana ability's Produced uses — pipes
	// and per-colour amounts included. "{G}" is Wild Growth, "{G}{G}"
	// is Overgrowth, "{W|U|B|R|G}" is Fertile Ground's "any color",
	// and prod.ProducedColorPipe() is "any type that land produced".
	//
	// Returning "" adds nothing, which is the printed answer whenever
	// the clause's input is not there (Utopia Sprawl before its colour
	// is chosen).
	Produced func(prod ManaProduced, source *Card, g *Game) string

	// ActiveWhen is the CR 716 / 719 / 721 / 709.5 designation gate,
	// exactly as on TriggeredAbility: this ability exists only while
	// its source has the designation named. Evaluated in
	// ManaTriggersForCard and nowhere else. The zero value is "no
	// gate", which is every card in the family today.
	ActiveWhen Designation
}

// CatalogManaTriggers returns the registered triggered mana abilities
// for a catalog key (one entry per effects.Spec.ManaTriggers element),
// or nil when the card has none. Set from CardDef at boot like every
// other per-slot hook; a test may stub it directly.
//
// Callers want ManaTriggersForCard, which applies the ability-removal
// key and the designation gate.
var CatalogManaTriggers func(key string) []ManaTrigger

// ManaTriggersForCard is the triggered mana abilities a permanent has
// right now: its catalog entry's, minus those a CR 613.1f
// ability-removing effect took away (CatalogAbilityKey) and minus
// those whose designation gate is unsatisfied.
//
// The same shape as TriggersForCard, and deliberately — see
// designations.go and ADR 0071. An Aura under Song of the Dryads has
// no key, so it has no mana triggers, and the land it enchants taps
// for exactly what it prints.
func ManaTriggersForCard(c Card) []ManaTrigger {
	if CatalogManaTriggers == nil {
		return nil
	}
	key := CatalogAbilityKey(c)
	if key == "" {
		return nil
	}
	return activeOnly(c, CatalogManaTriggers(key), func(t ManaTrigger) Designation {
		return t.ActiveWhen
	})
}

// fireManaTriggersLocked resolves every triggered mana ability the
// just-completed production satisfies (CR 605.1b), immediately and
// without the stack (CR 605.4a). The one firing function; see the file
// comment for the three sites that call it and the one that must not.
//
// `pending` is the auto-tap executor's still-unpaid colour
// requirements, and nil everywhere else. Non-nil makes a trigger whose
// output is a colour CHOICE pick greedily against the cast instead of
// queuing a prompt — the auto-tapper's contract is "no further player
// decisions", and a prompt appearing halfway through a cast would
// break it. See addTriggeredManaLocked.
//
// Order is battlefield order: deterministic, and the only thing the
// rules require of simultaneous mana abilities that all simply resolve.
//
// TWO PASSES, and the split is the CR 603.2 one: every condition is
// judged against the board as it was WHEN THE MANA WAS PRODUCED (pass
// one), and only then does anything resolve (pass two). It is also
// what makes the walk safe — adding mana emits events, an event runs
// listeners, and a listener that ever put a permanent on the
// battlefield would reallocate the slice the first pass is holding
// pointers into.
//
// A production that added nothing fires nothing: a land that was
// tapped for no mana was not "tapped for mana".
//
// Caller must hold g.mu in write mode.
func (g *Game) fireManaTriggersLocked(prod ManaProduced, pending *[]ColorRequirement) {
	if CatalogManaTriggers == nil || g.Battlefield == nil {
		return
	}
	if len(prod.Colors) == 0 || prod.Controller == uuid.Nil {
		return
	}
	type firing struct {
		source   uuid.UUID
		label    string
		produced string
	}
	var fired []firing
	for i := range g.Battlefield.Cards {
		source := &g.Battlefield.Cards[i]
		triggers := ManaTriggersForCard(*source)
		if len(triggers) == 0 {
			continue
		}
		for _, t := range triggers {
			if t.AppliesTo != nil && !t.AppliesTo(prod, source, g) {
				continue
			}
			produced := ""
			if t.Produced != nil {
				produced = t.Produced(prod, source, g)
			}
			if produced == "" {
				// "Could add mana" is part of what makes this a mana
				// ability at all (CR 605.1b); one that would add
				// nothing does nothing.
				continue
			}
			fired = append(fired, firing{source: source.InstanceID, label: t.Label, produced: produced})
		}
	}
	for _, f := range fired {
		g.addTriggeredManaLocked(prod.Controller, f.source, f.label, f.produced, pending)
	}
}

// addTriggeredManaLocked adds the mana a triggered mana ability
// produced. The token's Source is the TRIGGER's permanent (the Aura),
// not the land: the mana comes from Wild Growth's ability.
//
// It shares addManaSlotsLocked with AddManaForEffect rather than
// walking the slots itself, so a pipe from a trigger queues the same
// PendingChoiceMana a Birds of Paradise activation queues — or, with
// `pending` non-nil, takes the same greedy pick materializePlanLocked
// makes for the source's own slots.
//
// It does NOT fire mana triggers. That is the whole re-entrancy guard,
// and it is also the rule: nothing was tapped for this mana.
//
// Caller must hold g.mu in write mode.
func (g *Game) addTriggeredManaLocked(playerID, source uuid.UUID, label, produced string, pending *[]ColorRequirement) {
	p := g.playerByIDLocked(playerID)
	if p == nil || p.Eliminated {
		return
	}
	reason := func(slot ProducedManaEntry) string {
		if label != "" {
			return label
		}
		return addManaReason(slot)
	}
	if err := g.addManaSlotsLocked(p, source, produced, false, pending, reason); err != nil {
		g.EmitEvent(Event{
			Kind:     EventEffectError,
			Actor:    playerID,
			Source:   source,
			ErrorMsg: err.Error(),
		})
	}
}
