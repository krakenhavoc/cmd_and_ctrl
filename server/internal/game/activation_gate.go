package game

import (
	"errors"

	"github.com/google/uuid"
)

// activation_gate.go — #1210, ADR 0073's amendment of 2026-09-22: the
// one announce-time answer to "may this player activate this ability
// at all?" (CR 602.5, CR 101.2).
//
// It is the TWIN of cast_gate.go, deliberately and visibly: the same
// three shapes, the same two sources, the same "can't beats may"
// placement before any cost is paid. What it is not is the same
// FUNCTION. A cast has a spell, a source zone, a face and an
// announcement; an activation has a source object that stays where it
// is, an ability index inside that object, and a CR 605.1a mana /
// non-mana split no cast has. Widening CastQuery to carry both would
// have made every field optional and every restriction start by
// asking which of the two it was looking at, which is how Rule of Law
// comes to have an opinion about Pithing Needle.
//
// TWO SOURCES, and the split is the CR's own:
//
//  1. A STATIC ON A PERMANENT, board-wide. Cursed Totem, Linvala,
//     Collector Ouphe, Pithing Needle, Null Rod. Collected from the
//     battlefield through CatalogAbilityKey exactly as
//     CastRestrictionsForCard collects a cast restriction — so a
//     permanent that has lost its abilities (CR 613.1f) stops
//     restricting, and one whose designation gate is unsatisfied is
//     not there at all. Nothing is stored, which IS the duration: the
//     source leaving stops the restriction on the next query.
//
//  2. THE ABILITY'S OWN CONDITION, per ability. "Activate only if …"
//     / "Activate only during …" — ActivatedAbilityShape.Condition
//     and ManaAbilityShape.Condition, built by #743 and NOT moved
//     here. It is declared on the ability because it is printed on
//     the ability, it is checked by the two activation paths a few
//     lines below this gate's call, and it needs no board walk.
//
// What is deliberately NOT here:
//
//   - THE PER-PERMANENT BITS. CantActivate / CantActivateMana (Arrest,
//     Faith's Fetters) are a restriction on ONE permanent put there by
//     an Aura attached to it. They live on Characteristic.Restrictions
//     where layer 6 already answers for them, and both activation
//     paths check them where they always did. Folding them in would
//     have made every restriction walk the battlefield to answer a
//     question the layer engine had already answered.
//   - SPLIT SECOND. It restricts taking an ACTION, not activating a
//     particular ability, and both activation paths keep their own
//     check beside the gate call — the same line cast_gate.go draws.
//   - A RESTRICTION WITH A DURATION ("activated abilities can't be
//     activated this turn"). Same answer ADR 0073 §7 gave for
//     Silence: it wants #755's registries. A third source slots into
//     this function without changing its signature; that is the
//     extension point.

// ActivationAbility is the identity of the ability being activated —
// everything a restriction may ask about the ability itself, as
// against about its source or its activator.
//
// Two fields and not the whole shape, because the shape is two
// different types (ActivatedAbilityShape and ManaAbilityShape) and a
// gate that took either would have to be written twice. What a
// printed restriction reads is the label (for a log, and for a
// restriction keyed on a specific ability) and whether CR 605.1a
// makes this a mana ability.
type ActivationAbility struct {
	// Label is the ability's printed label — the same string the
	// activation record (#1181) is keyed by.
	Label string

	// Mana reports that this is a MANA ability (CR 605.1a): it
	// resolves without the stack and grants nobody priority.
	//
	// It rides the query rather than the call site, and that is the
	// one decision in this file worth arguing. It would have been a
	// line shorter to have ActivateManaAbility skip the gate —
	// except Pithing Needle exempts mana abilities ("…can't be
	// activated UNLESS THEY'RE MANA ABILITIES") and Cursed Totem
	// does not ("Activated abilities of creatures can't be
	// activated", full stop). The exemption is printed on the CARD,
	// so it belongs to the restriction; a call site that decided it
	// would make Cursed Totem unwritable without a second gate.
	Mana bool

	// SorcerySpeed is the ability's own "Activate only as a sorcery"
	// clause (CR 602.5d) — `ActivatedAbilityShape.SorcerySpeed`.
	//
	// Read by ActivationTimingOpenLocked (#1208) and by nothing in
	// this file: a BAN does not care how fast the ability is. It
	// rides this struct rather than a second identity type because
	// the two reads are asked about the same ability at the same
	// moment, and two spellings of "which ability" would be one
	// thing for the activation path to get out of step with itself.
	SorcerySpeed bool

	// Loyalty marks a loyalty ability (CR 606.1), which CR 606.3
	// makes sorcery-speed whatever the catalog entry says.
	//
	// The bool ADR 0073's #1210 scope note predicted — "a
	// restriction on LOYALTY abilities as a class … is a bool on
	// ActivationAbility, not a second gate" — arriving for the
	// timing read first. Derived from `AbilityCost.Loyalty != nil`
	// by ActivationAbilityOf, never set by a card file.
	Loyalty bool

	// Equip marks the CR 702.6 equip ability, because Leonin
	// Shikari's clause names the keyword: "you may activate EQUIP
	// abilities any time you could cast an instant".
	//
	// Set by EquipAbility (effects/attachments.go) and by nothing
	// else. It does NOT make equip a Spec field — attachments.go's
	// "equip is an ordinary activated ability" is unchanged; what is
	// new is that the keyword can be NAMED by a card that speaks
	// about it, which is a different claim.
	Equip bool
}

// ActivationAbilityOf is the one place an activated ability's shape
// becomes the identity the gate and the timing read ask about.
//
// A function rather than four struct literals, because the two
// derived fields are the ones a call site gets wrong: CR 606.3 rides
// the loyalty COST component and not a flag (so a catalog entry
// cannot forget it), and `Mana` is false here by construction — a
// mana ability is a ManaAbilityShape and takes the other entry point,
// which builds its own identity with `Mana: true`.
func ActivationAbilityOf(ab ActivatedAbilityShape) ActivationAbility {
	return ActivationAbility{
		Label:        ab.Label,
		SorcerySpeed: ab.SorcerySpeed,
		Loyalty:      ab.Cost.Loyalty != nil,
		Equip:        ab.Equip,
	}
}

// ActivationQuery is everything an activation restriction may look
// at. Passed by value for the reason CastQuery and CostQuery are: a
// restriction is consulted inside the activation path under the write
// lock, and handing it pointers into the battlefield slice would
// invite a card file to mutate the board while deciding whether an
// ability may be activated.
//
// Game is a pointer because a predicate may be board-wide. Treat it
// as READ-ONLY: a *ForEffect accessor is fine, a mutator is a bug and
// a public locking mutator is a deadlock.
type ActivationQuery struct {
	// Game is the game the activation is being announced in.
	// Read-only.
	Game *Game

	// Card is the OBJECT whose ability is being activated, as it
	// stands at the announce. Cursed Totem reads its types, Pithing
	// Needle its name.
	Card Card

	// Controller is the player activating the ability — the "you" of
	// the ACTIVATION. On the battlefield this is the source's
	// controller; off it (a cycling ability, #660) it is the card's
	// owner, which is CR 108.4.
	Controller uuid.UUID

	// Source is the permanent contributing the restriction. Its
	// Controller is the "you" of the ABILITY, as against Controller
	// above which is the "you" of the activation: Linvala restricts
	// everybody ELSE's creatures, Cursed Totem restricts everybody's
	// including its own controller's.
	Source Card

	// FromZone is the zone the activating object is in (CR 113.6) —
	// the battlefield for nearly every activation, a hand for
	// cycling. Present because "activated abilities of sources with
	// the chosen name" (Pithing Needle) says SOURCES and not
	// PERMANENTS, so a restriction that means the battlefield has to
	// say so.
	FromZone ZoneKind

	// Ability is which ability, and whether it is a mana one.
	Ability ActivationAbility
}

// ActivationRestriction is one "can't be activated" static ability
// contributed by a permanent on the battlefield.
//
// Declared as a struct of hooks rather than an interface for the
// reason StaticAbility, CastRestriction and CostModifier are: a card
// file writes a literal, not a type.
type ActivationRestriction struct {
	// Label is the clause as printed ("Activated abilities of
	// creatures can't be activated"). It is not decoration: a refused
	// activation returns it to the client, so the greyed row names
	// the card that said no rather than saying "invalid parameter".
	Label string

	// Forbids decides whether this restriction refuses THIS
	// activation. Nil forbids nothing, which makes an under-declared
	// card file a no-op rather than a lockout.
	//
	// Evaluated once per activation per restriction, under the write
	// lock. Read-only.
	Forbids func(q ActivationQuery) bool

	// ActiveWhen is the CR 716 / 719 / 721 designation gate, the same
	// field CostModifier and CastRestriction carry and evaluated in
	// the same accessor, so a gated-off restriction never reaches the
	// gate and the engine, the enumerator and the view cannot
	// disagree about whether it exists (ADR 0071).
	ActiveWhen Designation
}

// CatalogActivationRestrictions is the catalog hook the effects
// package wires at init, mirroring CatalogCastRestrictions. Nil, or a
// nil return, means the card restricts nobody's activations — which
// is nearly every card.
var CatalogActivationRestrictions func(oracleID string) []ActivationRestriction

// ActivationRestrictionsForCard returns the restrictions a permanent
// contributes right now: none for a permanent under an
// ability-removing effect (CatalogAbilityKey), and none for one whose
// designation gate is unsatisfied.
func ActivationRestrictionsForCard(c Card) []ActivationRestriction {
	if CatalogActivationRestrictions == nil {
		return nil
	}
	key := CatalogAbilityKey(c)
	if key == "" {
		return nil
	}
	return activeOnly(c, CatalogActivationRestrictions(key), func(r ActivationRestriction) Designation {
		return r.ActiveWhen
	})
}

// CantActivateError is what the gate returns: the sentinel plus the
// printed clause that refused the activation, so the client can say
// which card said no.
//
// Wraps ErrCantActivate, so every errors.Is(err, ErrCantActivate) in
// the tree — the Arrest and Faith's Fetters refusals, the client's
// error mapping — still matches and needed no change. The same shape
// CantCastError has over ErrCantCast.
type CantActivateError struct {
	// Reason is the printed clause ("Activated abilities of creatures
	// your opponents control can't be activated"), never an engine
	// phrase.
	Reason string

	// Source is the permanent whose static refused the activation.
	Source uuid.UUID
}

func (e *CantActivateError) Error() string {
	if e.Reason == "" {
		return "game: an effect prevents activating this ability"
	}
	return "game: " + e.Reason
}

// Unwrap lets errors.Is(err, ErrCantActivate) match a
// *CantActivateError.
func (e *CantActivateError) Unwrap() error { return ErrCantActivate }

// ActivationGateLocked is CR 101.2's "can't beats may" asked of an
// ACTIVATION, once, at the announce. It reports nil when `activator`
// may activate `ability` of `card` out of `zone` right now, and a
// *CantActivateError naming the clause that refused it otherwise.
//
// FOUR CALLERS, ONE FUNCTION, and that is the entire point:
//
//   - ActivateCatalogAbility, after the CR 113.6 zone check (so the
//     ability is the one the view published) and BEFORE the timing
//     check, X, the targets and every cost. A refused activation
//     costs nothing.
//   - ActivateManaAbility, at the same point relative to its own
//     gates — after the ability is resolved, before the exhaust
//     record and the condition.
//   - legal.abilityMovesForSource and legal.manaMoves, so a bot is
//     never offered an activation the engine refuses (#544).
//   - gatherTapSources (autotap.go), which is the caller the CAST
//     gate has no equivalent of: the auto-tapper never goes through
//     ActivateManaAbility — materializePlanLocked taps the permanent
//     and mints its mana directly — so a plan built without asking
//     would tap a Birds of Paradise under a Cursed Totem and produce
//     mana the rule forbids.
//
// And one view stamp, ActivatedAbilityView / ManaAbilityView's
// `cant_activate`, so the client greys the row and says why from
// server data rather than from a rule it reimplemented.
//
// ProducibleManaLocked (CR 106.7) deliberately does NOT ask: "could
// produce" is a question about what the ability would do IF IT
// RESOLVED, not about whether it can be activated, so a Reflecting
// Pool beside a Cursed-Totem'd Birds of Paradise still sees {G}.
// #1183's exhaust narrowing is declared in that file as the one
// exception and this is not a second one.
//
// Caller must hold g.mu (read or write).
func (g *Game) ActivationGateLocked(activator uuid.UUID, card Card, zone ZoneKind, ability ActivationAbility) error {
	if g == nil || g.Battlefield == nil || CatalogActivationRestrictions == nil {
		return nil
	}
	q := ActivationQuery{
		Game:       g,
		Card:       card,
		Controller: activator,
		FromZone:   zone,
		Ability:    ability,
	}
	for i := range g.Battlefield.Cards {
		src := g.Battlefield.Cards[i]
		for _, r := range ActivationRestrictionsForCard(src) {
			if r.Forbids == nil {
				continue
			}
			q.Source = src
			if r.Forbids(q) {
				return &CantActivateError{Reason: r.Label, Source: src.InstanceID}
			}
		}
	}
	return nil
}

// AnyActivationRestrictionsForEffect reports whether anything on the
// battlefield restricts activating at all. The fast negative the view
// and the enumerator take before walking a board permanent by
// permanent and ability by ability — in almost every game nothing
// restricts anything, and the gate is otherwise pure cost on a full
// battlefield.
//
// Caller must hold g.mu (read or write).
func (g *Game) AnyActivationRestrictionsForEffect() bool {
	if g == nil || g.Battlefield == nil || CatalogActivationRestrictions == nil {
		return false
	}
	for i := range g.Battlefield.Cards {
		if len(ActivationRestrictionsForCard(g.Battlefield.Cards[i])) > 0 {
			return true
		}
	}
	return false
}

// CantActivateReasonLocked is ActivationGateLocked's answer as a
// STRING, for the two view stamps: the printed clause that refuses
// this activation, or "" when nothing does.
//
// A helper rather than an errors.As at each stamp site because the
// view asks the question per ability row and both callers want the
// same fallback for the unreachable case (Register refuses a
// restriction with no label, exactly as it does for a cast one).
//
// Caller must hold g's read lock.
func (g *Game) CantActivateReasonLocked(activator uuid.UUID, card Card, zone ZoneKind, ability ActivationAbility) string {
	err := g.ActivationGateLocked(activator, card, zone, ability)
	if err == nil {
		return ""
	}
	var cant *CantActivateError
	if errors.As(err, &cant) && cant.Reason != "" {
		return cant.Reason
	}
	return "An effect prevents activating this ability."
}
