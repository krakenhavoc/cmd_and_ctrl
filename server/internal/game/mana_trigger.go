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
//
// # The judge and the executor (ADR 0074, amendment of 2026-10-30)
//
// Firing is two passes, and they are two functions. manaTriggerFiringsLocked
// is the JUDGE: it decides which triggered mana abilities a production
// satisfies and changes nothing, so it is safe under a read lock and on
// a production that has not happened yet. fireManaTriggersLocked is the
// EXECUTOR: it calls the judge and then resolves what the judge found.
// The auto-tap planner is the judge's second reader (#4716,
// autotap_trigger_mana.go), so the plan and the payment cannot disagree
// about what a trigger adds: there is one copy of each trigger's
// condition. manaTriggersExistLocked is the cheap gate in front of both.

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
// Both callbacks run under g.mu in EITHER mode: write when the executor
// fires a real production, read when the judge (manaTriggerFiringsLocked)
// is asked about one — the auto-tap planner asks it about a production
// that has not happened yet (ADR 0074, amendment of 2026-10-30). So they
// are READ-ONLY, and the rule carries weight:
//
//   - read the tapped permanent through prod.Source, never through a
//     fresh battlefield lookup (the production may not have happened);
//   - read nothing from the mana pool;
//   - write nothing, and use *ForEffect accessors, never a public
//     locking mutator.
//
// TestEveryCatalogManaTriggerIsReadOnly (cards/effects) checks every
// catalogued trigger against it.
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

	// TappedSkipsNextUntap is Winter's Night's rider (#1997): "That land
	// doesn't untap during its controller's next untap step." When the
	// ability resolves, the tapped permanent gets the ordinary
	// next-untap marker for the player who controls it then (CR 502.3,
	// SkipNextUntapForEffect), after its mana is added. Plain data, so
	// the rider adds no closure route. False for every other card.
	TappedSkipsNextUntap bool
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
func ManaTriggersForCard(c Card) []ManaTrigger { return manaTriggersOf(&c) }

// manaTriggersOf is ManaTriggersForCard without the copy: Card is over
// a kilobyte and this is asked per permanent per walk (#1498).
func manaTriggersOf(c *Card) []ManaTrigger {
	if CatalogManaTriggers == nil {
		return nil
	}
	key := catalogAbilityKeyOf(c)
	if key == "" {
		return nil
	}
	return activeOnly(c, CatalogManaTriggers(key), func(t ManaTrigger) Designation {
		return t.ActiveWhen
	})
}

// manaTriggerFiring is one triggered mana ability the judge found
// satisfied by a production: whose it is, what the log calls it and the
// mana it adds. Plain data, with no pointer into the battlefield, so
// the executor can resolve a list while resolving moves the board.
type manaTriggerFiring struct {
	// source is the permanent carrying the ability (the Aura, the Mana
	// Flare), or the spell that set up a delayed record. The mana the
	// ability adds is attributed to it.
	source uuid.UUID
	label  string
	// produced is the mana the ability adds, in the ParseProducedMana
	// grammar. Never empty: an ability that would add nothing is not a
	// firing (CR 605.1b "could add mana").
	produced string
	// skip is ManaTrigger.TappedSkipsNextUntap (#1997).
	skip bool
}

// addsMana reports whether p is a production that can trigger anything:
// a land that was tapped for no mana was not "tapped for mana", and mana
// that landed in nobody's pool triggers nothing.
func (p ManaProduced) addsMana() bool {
	return len(p.Colors) > 0 && p.Controller != uuid.Nil
}

// manaTriggersExistLocked reports whether ANY triggered mana ability
// could fire: a permanent on the battlefield with a mana trigger right
// now, or a delayed mana record a spell set up (Bubbling Muck). It is
// the cheap gate in front of manaTriggerFiringsLocked: on nearly every
// board the answer is no, and a caller that would otherwise build a
// production to ask about skips all of it.
//
// A false here means manaTriggerFiringsLocked returns nothing for every
// production. Deliberately asks only whether a trigger exists and not
// AppliesTo, the same pre-filter producesManaReplacementsExistLocked
// uses: a true means the judge is worth asking.
//
// Caller must hold g.mu (a read lock is enough).
func (g *Game) manaTriggersExistLocked() bool {
	for _, dt := range g.DelayedTriggers {
		if dt != nil && dt.isManaTrigger() {
			return true
		}
	}
	if g.Battlefield == nil || CatalogManaTriggers == nil {
		return false
	}
	for i := range g.Battlefield.Cards {
		if len(manaTriggersOf(&g.Battlefield.Cards[i])) > 0 {
			return true
		}
	}
	return false
}

// manaTriggerFiringsLocked is the JUDGE (pass one of the two passes
// fireManaTriggersLocked is made of): which triggered mana abilities
// does `prod` satisfy (CR 605.1b)? It walks the delayed records a spell
// set up for the turn (High Tide, Bubbling Muck, Molten Tide) and then
// the battlefield, asks each ManaTrigger's AppliesTo and Produced, and
// returns the firings in the order they resolve: scheduling order, then
// battlefield order. That order is deterministic, and it is all the
// rules require of mana abilities that simply resolve.
//
// It CHANGES NOTHING. `prod` need not be a production that has happened:
// the auto-tap planner asks about the one it expects the executor to
// make, under a read lock. Every condition is judged against the board
// as it stands at the call. For the executor that is the board as it was
// WHEN THE MANA WAS PRODUCED (CR 603.2), because nothing has resolved
// yet: resolving is pass two. Judging first is also what makes the walk
// safe. Adding mana emits events, an event runs listeners, and a
// listener that ever put a permanent on the battlefield would reallocate
// the slice this walk is holding pointers into.
//
// A production that adds nothing fires nothing: a land that was tapped
// for no mana was not "tapped for mana". An ability whose Produced is ""
// is not a firing either.
//
// Caller must hold g.mu (a read lock is enough).
func (g *Game) manaTriggerFiringsLocked(prod ManaProduced) []manaTriggerFiring {
	if !prod.addsMana() {
		return nil
	}
	var fired []manaTriggerFiring
	// CR 605.1b / 603.7b: triggered mana abilities a spell set up for
	// the turn (Bubbling Muck). Not tied to a permanent, so they are
	// read from the delayed queue, in scheduling order.
	for _, dt := range g.DelayedTriggers {
		if dt == nil || !dt.isManaTrigger() || !prod.Source.HasSubtype(dt.ManaTapSubtype) {
			continue
		}
		// "Whenever YOU tap" (Pyre Rhymer // Molten Tide): an opponent's
		// land does not trigger it, and it does not hand them the mana.
		if dt.ManaTapControllerOnly && dt.Controller != prod.Controller {
			continue
		}
		fired = append(fired, manaTriggerFiring{source: dt.SourceCardID, label: dt.Label, produced: dt.ManaAdds})
	}
	for i := 0; g.Battlefield != nil && CatalogManaTriggers != nil && i < len(g.Battlefield.Cards); i++ {
		source := &g.Battlefield.Cards[i]
		triggers := manaTriggersOf(source)
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
			fired = append(fired, manaTriggerFiring{
				source:   source.InstanceID,
				label:    t.Label,
				produced: produced,
				skip:     t.TappedSkipsNextUntap,
			})
		}
	}
	return fired
}

// ManaTriggerSourcesForTest is the judge for a test in another package
// (cards/effects' TestEveryCatalogManaTriggerIsReadOnly): the source of
// each triggered mana ability `prod` satisfies, in the order the executor
// would resolve them. The caller holds g.mu in either mode, normally by
// calling this inside Game.ReadSnapshot, which is the point of the test.
func (g *Game) ManaTriggerSourcesForTest(prod ManaProduced) []uuid.UUID {
	var out []uuid.UUID
	for _, f := range g.manaTriggerFiringsLocked(prod) {
		out = append(out, f.source)
	}
	return out
}

// fireManaTriggersLocked is the EXECUTOR: it resolves every triggered
// mana ability the just-completed production satisfies (CR 605.1b),
// immediately and without the stack (CR 605.4a). It asks
// manaTriggerFiringsLocked which those are (pass one, every condition
// judged against the board as it was when the mana was produced, CR
// 603.2) and then resolves them (pass two). The one firing function; see
// the file comment for the three sites that call it and the one that
// must not.
//
// `pending` is the auto-tap executor's still-unpaid colour
// requirements, and nil everywhere else. Non-nil makes a trigger whose
// output is a colour CHOICE pick greedily against the cast instead of
// queuing a prompt — the auto-tapper's contract is "no further player
// decisions", and a prompt appearing halfway through a cast would
// break it. See addTriggeredManaLocked.
//
// `booked` is the colour the plan booked for each mana these firings add,
// in firing order (plannedTap.TriggerColors, ADR 0074 amendment of
// 2026-10-30), and nil everywhere but the auto-tap executor. A pick makes
// its booked colour instead of the greedy one, so the trigger cannot take
// a pip a later source in the plan was booked for. It is used only when
// it names one colour per mana slot of these firings; otherwise the plan
// was made for a different board and every pick is greedy, as before.
//
// Order is the judge's: simultaneous mana abilities that all simply
// resolve have no other order the rules ask for.
//
// A production that added nothing fires nothing: a land that was
// tapped for no mana was not "tapped for mana".
//
// Caller must hold g.mu in write mode.
func (g *Game) fireManaTriggersLocked(prod ManaProduced, pending *[]ColorRequirement, booked []string) {
	// The gate goes first. On a board with no mana trigger the judge
	// would walk every permanent to say nothing. The gate walks the same
	// permanents once and stops at the first trigger it finds, so such a
	// board pays what it paid before.
	if !prod.addsMana() || !g.manaTriggersExistLocked() {
		return
	}
	firings := g.manaTriggerFiringsLocked(prod)
	if booked != nil {
		n := 0
		for _, f := range firings {
			n += producedSlotCount(f.produced)
		}
		if n != len(booked) {
			booked = nil
		}
	}
	for _, f := range firings {
		var mine []string
		if booked != nil {
			n := producedSlotCount(f.produced)
			mine, booked = booked[:n], booked[n:]
		}
		g.addTriggeredManaLocked(prod.Controller, f.source, f.label, f.produced, pending, mine)
		if f.skip {
			// #1997: the tapped permanent, if it is still there, skips
			// its controller's next untap step.
			if land := g.findBattlefieldCardLocked(prod.SourceID()); land != nil {
				_ = g.SkipNextUntapForEffect(land.InstanceID, land.Controller)
			}
		}
	}
}

// producedSlotCount is how many slots a produced-mana string parses to,
// the unit plannedTap.TriggerColors counts in; 0 for one that does not
// parse, which adds nothing.
func producedSlotCount(produced string) int {
	slots, err := ParseProducedMana(produced)
	if err != nil {
		return 0
	}
	return len(slots)
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
// `booked` is the plan's colour for each slot of `produced`, or nil (see
// fireManaTriggersLocked).
//
// Caller must hold g.mu in write mode.
func (g *Game) addTriggeredManaLocked(playerID, source uuid.UUID, label, produced string, pending *[]ColorRequirement, booked []string) {
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
	if err := g.addManaSlotsLocked(p, source, produced, false, nil, nil, pending, booked, reason); err != nil {
		g.EmitEvent(Event{
			Kind:     EventEffectError,
			Actor:    playerID,
			Source:   source,
			ErrorMsg: err.Error(),
		})
	}
}
