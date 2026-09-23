package game

import "github.com/google/uuid"

// designations.go is the engine's ONE gate on a printed ability —
// ADR 0071. A DESIGNATION is a marker a permanent has on the
// battlefield that is not a counter, not a characteristic and not
// copiable, and whose only job is to switch some of the permanent's
// own printed abilities on:
//
//	CR 716.2   a Class's LEVEL      "as long as this Class is level N or greater, it has …"
//	CR 719.3   a Case being SOLVED  "Solved — [ability]"
//	CR 721.2   a station THRESHOLD  "as long as this has N or more charge counters, it has …"
//	CR 709.5   a Room's UNLOCKED    a locked door has no rules text at all
//	           DOOR                 (designed here, built by #886)
//
// Four printed mechanics, one sentence. So one field, one predicate,
// one place — and nowhere else in the engine asks the question.
//
// # Where the gate is evaluated, and why it is only there
//
// ADR 0046 made `CatalogAbilityKey` the single question "what
// abilities does this object have right now", because the CR 613.1f
// bug it fixed was structural: the catalog was consulted at USE time,
// from wherever the use happened, so a seventh caller could
// reintroduce it for free. ADR 0069 put face-down suppression at
// `CatalogKey`, the one place a Card becomes a catalog key, for the
// same reason.
//
// A designation gate is that question again — it is "does this object
// have this ability right now" with a different reason for the answer
// — so it is evaluated at the same seam: the four accessors below,
// which sit directly on top of those two key functions and are the
// only way the engine turns an OBJECT into the abilities it has.
//
//	StaticAbilitiesForCard     the layer pass, the emblem gather, the hand-size probe
//	TriggersForCard            the trigger harvester, the LTB harvest, the
//	                           simultaneous-exit harvest, #925's declared-zone
//	                           walk (trigger_zones.go), the Saga final chapter
//	ActivatedAbilitiesForCard  activation, the legal enumerator, the view  (activated.go)
//	CostModifiersForCard       the cast pricer                            (cost_modifier.go)
//
// An inactive ability is simply NOT IN THE LIST the object hands
// back. There is no second filter anywhere — the layer pass never
// sees a level-3 anthem on a level-1 Class, the harvester never
// matches a solved Case's trigger before it is solved, the enumerator
// never offers a bot an ability the engine would refuse (the #544
// invariant), and the wire never renders one.
//
// # What the gate deliberately does NOT do
//
// It does not know about zones, the stack, or a spell being cast. A
// designation is battlefield state (CR 400.7 clears all of it), and
// off the battlefield every gate but DesignationAlways is false by
// construction — a Class in a graveyard is level 1 with none of its
// later abilities, which is CR 113.6 falling out for free.
//
// It does not re-check at resolution. An activated ability that was
// legal when it was announced resolves even if the permanent lost the
// designation in response (CR 602.1b: an activation instruction is
// not part of the effect), and a trigger that fired stays on the
// stack (CR 603.2). Both are the existing behaviour for every other
// reason an ability can stop existing, and neither needed code here.
//
// # The per-slot Catalog* hooks stay
//
// Two dozen test files stub CatalogStaticAbilities / CatalogTriggers /
// CatalogActivatedAbilities individually to inject catalog behaviour
// without importing the effects package (carddef.go's header has the
// history). The accessors below read THROUGH those hooks, so a stub
// still works and the gate still applies to what the stub returns.

// DesignationKind is which designation an ability's gate names.
type DesignationKind uint8

const (
	// DesignationAlways is the zero value: no gate, the ability is
	// always present. Every ability in the catalog that does not
	// declare otherwise.
	DesignationAlways DesignationKind = iota

	// DesignationClassLevel is CR 716.2's "as long as this Class is
	// level N or greater". N is the level the ability is printed
	// against.
	DesignationClassLevel

	// DesignationCaseSolved is CR 719.3's "Solved — [ability]".
	DesignationCaseSolved

	// DesignationChargeCounters is CR 721.2a's "{N+}": as long as
	// this permanent has N or more charge counters. N is the
	// threshold printed to the left of the bar.
	//
	// It reads the count LIVE, which is the one behavioural
	// difference from a Saga's lore ratchet: a Spacecraft that loses
	// charge counters loses the abilities again, because CR 721.2a is
	// a continuous "as long as", not an event.
	DesignationChargeCounters

	// DesignationDoorUnlocked is CR 709.5's locked half — an ability
	// printed on a Room's door exists only while that door is
	// unlocked.
	//
	// RESERVED, NOT BUILT (ADR 0071 decision 3). Card has no unlocked
	// state yet, so Active is false for it, no effects constructor
	// produces one, and TestNoRegisteredSpecDeclaresADoorGate fails
	// the build if a card file tries. #886 adds the state and the
	// unlock special action; nothing else here changes when it does.
	DesignationDoorUnlocked
)

// DoorSide names which half of a Room a DesignationDoorUnlocked gate
// is printed on (CR 709.5a). Meaningless for every other kind.
type DoorSide uint8

const (
	// DoorNone is the zero value — not a door gate.
	DoorNone DoorSide = iota
	DoorLeft
	DoorRight
)

// Designation is the gate on one printed ability. The zero value is
// "no gate", which is what every ability in the catalog that does not
// declare one gets, so the field costs existing cards nothing.
type Designation struct {
	// Kind is which designation this gate names.
	Kind DesignationKind

	// N is the threshold for the two counted kinds — the class level
	// (CR 716.2) or the charge-counter count (CR 721.2). Both are
	// "N or greater", never "exactly N": a level-3 Class keeps its
	// level-2 abilities, and a Spacecraft at 9 charge counters has
	// everything printed at 7+.
	N int

	// Door is which half of a Room a DoorUnlocked gate is printed on.
	Door DoorSide
}

// Active reports whether this gate is satisfied by the object right
// now — THE predicate, and the only place any of the four rules is
// read.
//
// It takes a Card and nothing else. No *Game, no lock, no zone: that
// is what lets it be called from inside the layer pass, whose
// AppliesTo predicates must stay passive or the recompute recurses
// (see the note on Card.Effective in card.go), and from the trigger
// harvester, which runs it once per declared ability per battlefield
// card per event.
func (d Designation) Active(c Card) bool {
	switch d.Kind {
	case DesignationAlways:
		return true
	case DesignationClassLevel:
		return ClassLevelOf(c) >= d.N
	case DesignationCaseSolved:
		return c.Solved
	case DesignationChargeCounters:
		return c.Counters[CounterCharge] >= d.N
	case DesignationDoorUnlocked:
		// Reserved. Nothing can unlock a door yet, so nothing is
		// unlocked — and no registered card declares this gate, so
		// the false is unobservable rather than a silent nerf.
		return false
	}
	return false
}

// IsGate reports whether this designation actually gates anything —
// false for the zero value. Used by the ability filters to skip the
// copy entirely for the overwhelming majority of cards.
func (d Designation) IsGate() bool { return d.Kind != DesignationAlways }

// ClassLevel builds a CR 716.2 gate: the ability exists while the
// Class is level n or greater.
func ClassLevel(n int) Designation {
	return Designation{Kind: DesignationClassLevel, N: n}
}

// CaseSolved builds a CR 719.3 gate: the ability exists while the
// Case is solved.
func CaseSolved() Designation { return Designation{Kind: DesignationCaseSolved} }

// ChargeCounters builds a CR 721.2a "{n+}" gate: the ability exists
// while the permanent has n or more charge counters.
func ChargeCounters(n int) Designation {
	return Designation{Kind: DesignationChargeCounters, N: n}
}

// ClassLevelOf is the permanent's current Class level (CR 716.2b): a
// Class permanent with no level designation is level 1, so the zero
// value of Card.ClassLevel reads as 1.
//
// Answers 1 for every permanent in the game, Class or not, which is
// harmless: only an ability that declares a ClassLevel gate ever asks,
// and only a Class prints one.
func ClassLevelOf(c Card) int {
	if c.ClassLevel < 1 {
		return 1
	}
	return c.ClassLevel
}

// ClassSubtype is the subtype that makes an enchantment a Class
// (CR 716.1), and CaseSubtype the one that makes it a Case
// (CR 719.1). Lowercase, like SagaSubtype: HasSubtype folds case.
const (
	ClassSubtype = "class"
	CaseSubtype  = "case"
)

// IsClass and IsCase read the EFFECTIVE subtypes, so a permanent an
// effect turned into a Class counts and one whose types were
// overwritten does not — the same reading IsSaga takes, and for the
// same reason.
//
// They exist for the WIRE, not for the gate: a designation gate asks
// about the level or the flag, never about the type line, so a Class
// that stopped being a Class keeps whatever level it had. What these
// answer is "should the client show a level badge on this card",
// which is a question about what the permanent is right now.
func IsClass(c Card) bool { return c.HasSubtype(ClassSubtype) }
func IsCase(c Card) bool  { return c.HasSubtype(CaseSubtype) }

// activeOnly drops the entries whose gate the object does not satisfy
// — THE filter, shared by all four accessors.
//
// Returns the input slice unchanged when nothing is gated, which is
// every card in the catalog but a handful, so the hot paths allocate
// nothing. The catalog's slices are built once at Register and never
// mutated, so handing the original back is safe (see CardDef).
func activeOnly[T any](c Card, all []T, gate func(T) Designation) []T {
	gated := false
	for i := range all {
		if gate(all[i]).IsGate() {
			gated = true
			break
		}
	}
	if !gated {
		return all
	}
	out := make([]T, 0, len(all))
	for i := range all {
		if gate(all[i]).Active(c) {
			out = append(out, all[i])
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// StaticAbilitiesForCard is the static abilities a permanent
// contributes to the layer engine right now: its catalog entry's
// statics, minus those whose designation gate is unsatisfied.
//
// CatalogKey, not CatalogAbilityKey, and deliberately: the layer pass
// resets every effective characteristic to printed before it gathers,
// so there is nothing for the removal accessor to read yet — see the
// note in activeStaticAbilitiesLocked. CR 613.1f silencing happens
// per bucket, in applyBucketLocked, and this gate is orthogonal to it.
func StaticAbilitiesForCard(c Card) []StaticAbility {
	if CatalogStaticAbilities == nil {
		return nil
	}
	key := CatalogKey(c)
	if key == "" {
		return nil
	}
	return activeOnly(c, CatalogStaticAbilities(key), func(a StaticAbility) Designation {
		return a.ActiveWhen
	})
}

// TriggersForCard is the triggered abilities a permanent has right
// now: its catalog entry's triggers, minus those a CR 613.1f
// ability-removing effect took away (CatalogAbilityKey) and minus
// those whose designation gate is unsatisfied.
//
// Off the battlefield this degrades to the printed list, which is
// what the LKI harvest and the from-stack cascade scan want.
func TriggersForCard(c Card) []TriggeredAbility {
	// CR 702.168a / CR 701.58a, ADR 0069 decision 3 and ADR 0082
	// decision 8: a DISGUISED or CLOAKED object has ward {2}, and it
	// is the one ability a CR 708.2 object has. Answered before the
	// catalog read rather than folded into it, because the catalog
	// read is exactly what CR 708.2a silences — this ability belongs
	// to the face-down STATE, not to the card underneath.
	if ward := faceDownWardLocked(c); len(ward) > 0 {
		return ward
	}
	if CatalogTriggers == nil {
		return nil
	}
	key := CatalogAbilityKey(c)
	if key == "" {
		return nil
	}
	return activeOnly(c, CatalogTriggers(key), func(t TriggeredAbility) Designation {
		return t.ActiveWhen
	})
}

// TriggersForKey is TriggersForCard for a harvest that holds a
// last-known-information snapshot rather than a live card: the LTB
// path, which has already decided its own key and already read
// AbilitiesRemoved off the snapshot (CR 603.10).
//
// The gate still applies, and it applies to the SNAPSHOT: a Case that
// was solved when it died has its solved dies-trigger, and one that
// was not does not. That is the same rule the LKI check next to it
// enforces for ability removal, read off the same object.
func TriggersForKey(key string, lki Card) []TriggeredAbility {
	if CatalogTriggers == nil || key == "" {
		return nil
	}
	return activeOnly(lki, CatalogTriggers(key), func(t TriggeredAbility) Designation {
		return t.ActiveWhen
	})
}

// CostModifiersForCard is the "spells cost {N} more / less" statics a
// permanent contributes right now, gate applied. Fortune Teller's
// Talent's level-3 reduction is the card that needs it.
func CostModifiersForCard(c Card) []CostModifier {
	if CatalogCostModifiers == nil {
		return nil
	}
	key := CatalogAbilityKey(c)
	if key == "" {
		return nil
	}
	return activeOnly(c, CatalogCostModifiers(key), func(m CostModifier) Designation {
		return m.ActiveWhen
	})
}

// --- the designations themselves ------------------------------------

// SetClassLevelForEffect sets a Class permanent's level (CR 716.2)
// and announces it. The one writer of Card.ClassLevel outside the
// snapshot restore.
//
// The level is a DESIGNATION, not a counter: it is not in
// Card.Counters, so nothing proliferates it, doubles it, or pays a
// counter-removal cost with it, and none of those rules needed a
// clause here to say so.
//
// Levels only ever go UP on a printed card ("this Class's level
// becomes N", activated only at N-1), but nothing here enforces a
// direction — an effect that set a lower level would be honoured,
// which is the conservative reading of a rule no card exercises.
//
// Caller must hold g.mu in write mode.
func (g *Game) SetClassLevelForEffect(cardID uuid.UUID, level int) error {
	card := findBattlefieldCard(g, cardID)
	if card == nil {
		return ErrCardNotFound
	}
	if level < 1 {
		return ErrInvalidParam
	}
	if card.ClassLevel == level {
		return nil
	}
	card.ClassLevel = level
	g.EmitEvent(Event{
		Kind:   EventClassLevel,
		Actor:  card.Controller,
		Source: cardID,
		CardID: cardID,
		Target: cardID,
		Amount: level,
	})
	return nil
}

// SolveCaseForEffect marks a Case solved (CR 719.3) and announces it.
//
// Idempotent, and that matters: a solved Case STAYS solved while it
// is on the battlefield (CR 719.3b), so a second "to solve" trigger
// that somehow resolves must not re-announce and must not re-fire
// whatever watches EventCaseSolved.
//
// Caller must hold g.mu in write mode.
func (g *Game) SolveCaseForEffect(cardID uuid.UUID) error {
	card := findBattlefieldCard(g, cardID)
	if card == nil {
		return ErrCardNotFound
	}
	if card.Solved {
		return nil
	}
	card.Solved = true
	g.EmitEvent(Event{
		Kind:   EventCaseSolved,
		Actor:  card.Controller,
		Source: cardID,
		CardID: cardID,
		Target: cardID,
	})
	return nil
}

// IsSolved reports whether the named battlefield permanent is a
// solved Case. False for anything not on the battlefield, which is
// CR 400.7 — a Case that left is a new object and is not solved.
//
// Caller must hold g.mu.
func (g *Game) IsSolved(cardID uuid.UUID) bool {
	card := findBattlefieldCard(g, cardID)
	return card != nil && card.Solved
}

// ClassLevelFor is the current level of the named battlefield
// permanent, or 0 when it is not on the battlefield. Read by the
// level-up ability's CR 716.2e condition and by the wire view.
//
// Zero rather than 1 for "not found" on purpose: the caller asking is
// asking about a permanent, and "there is no permanent" is not
// "level 1".
//
// Caller must hold g.mu.
func (g *Game) ClassLevelFor(cardID uuid.UUID) int {
	card := findBattlefieldCard(g, cardID)
	if card == nil {
		return 0
	}
	return ClassLevelOf(*card)
}
