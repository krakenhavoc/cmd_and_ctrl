package game

import "github.com/google/uuid"

// trigger_suppression.go — "[an event] doesn't cause abilities to
// trigger" (#1735, ADR 0018 amendment 2026-10-01).
//
// Torpor Orb: "Creatures entering don't cause abilities to trigger."
// Hushbringer adds "or dying". The third paragraph of Elesh Norn,
// Mother of Machines narrows it to "abilities of permanents your
// opponents control". Each one is a static ability that changes
// whether an event TRIGGERS an ability at all (CR 603.2). It is not a
// replacement effect (CR 614): the event still happens, the permanent
// still enters, and the ability simply never triggers. So there is
// no instance to double, no "you may" to ask, no target to choose,
// nothing on the stack, and nothing for TurnTally to count.
//
// A suppressor is the TriggerDoubler's sibling and is read at the same
// point of the same harvest: harvestMatchLocked asks the suppressors
// first, and a suppressed match returns before the OncePerBatch guard
// and before any doubler is consulted. That order is the composition
// rule. A suppressed ability is not doubled (Elesh Norn doubles an
// ability that a permanent entering "causes to trigger", and this one
// did not trigger). And a "whenever one or more" ability does not
// spend its once-per-batch slot on an event that did not trigger it.
//
// Which board is asked is the one rule this file adds, and it follows
// CR 603.10:
//
//   - Away from a leaves-the-battlefield event, the game checks the
//     board immediately AFTER the event, with the continuous effects
//     that exist then. A suppressor that enters with the permanent
//     (including the suppressor's own entry) applies. A suppressor
//     that has already left the battlefield does not apply, including
//     one that left in the same simultaneous exit
//     (modifierCandidate.live).
//   - A leaves-the-battlefield ability "looks back in time" (CR
//     603.10a), so it uses the abilities that existed immediately
//     BEFORE the event. A Hushbringer that dies in the same wipe as
//     the creatures still stops their dies triggers, and it stops the
//     triggers of its own death too. The candidates are the ones the
//     doubler already reads for that event: the open simultaneous-exit
//     batch and the leaving permanent's last-known information.
//
// Elesh Norn's two paragraphs are a doubler and a suppressor. They use
// the same definition of "entering" (effects.isEntering), so the two
// halves of one card can't disagree about what entered.

// TriggerSuppressor is a catalog-declared static saying that an event
// doesn't cause a triggered ability to trigger. Declare one on
// Spec.TriggerSuppressors with the effects-package helpers
// (SuppressesEntering, SuppressesDying, OfOpponentsPermanents).
type TriggerSuppressor struct {
	// Label is the card's printed name. It is used only in logs and
	// tests: a suppressed ability leaves nothing on the wire to
	// attribute.
	Label string

	// Suppresses reports whether the ability described by q does not
	// trigger from q.Event. It is called under g.mu in write mode, from
	// the harvest. It must not mutate state or take a public lock.
	Suppresses func(g *Game, q TriggerSuppressionQuery) bool

	// ActiveWhen is the ADR 0071 designation gate, which works exactly
	// as TriggerDoubler.ActiveWhen does: it is checked against the
	// SUPPRESSOR's own card. The zero value means "always active".
	ActiveWhen Designation
}

// TriggerSuppressionQuery is what a suppressor is asked. It is
// TriggerDoublingQuery with two changes: the static's own card is the
// Suppressor rather than the Doubler, and SourceIsPermanent replaces
// FromSpell, because the suppressor cards' wording ("abilities of
// permanents your opponents control") is about the object the ability
// belongs to.
type TriggerSuppressionQuery struct {
	Event Event

	// Suppressor is the permanent the static is printed on, and
	// SuppressorLKI is its characteristics as the harvest reads them.
	// For a dying event that can be last-known information (see the
	// file comment).
	Suppressor    Card
	SuppressorLKI Characteristic

	// Source is the object whose ability would trigger. SourceLKI is
	// its characteristics as the harvest reads them, so
	// SourceLKI.Controller is the controller of the permanent whose
	// ability it is.
	Source    Card
	SourceLKI Characteristic

	// SourceIsPermanent is true when the ability belongs to a
	// permanent: one on the battlefield, or (for a leaves-the-
	// battlefield ability) the permanent that just left. It is false for
	// a spell's "when you cast this spell", an emblem's ability, and a
	// card's ability that works from a hand, graveyard or exile (CR
	// 113.6).
	SourceIsPermanent bool

	// Ability is the triggered ability. It is nil for an
	// event-conditioned delayed trigger (CR 603.7), which has no catalog
	// declaration.
	Ability *TriggeredAbility

	// Subject is the object the event is about: the permanent that
	// entered, or the one that left with its last-known
	// characteristics. This is the same subject the doubler is given.
	Subject    uuid.UUID
	SubjectLKI Characteristic
	HasSubject bool
}

// CatalogTriggerSuppressors returns the suppressors a catalog card
// declares. carddef.go sets it from CardDef.TriggerSuppressors. It is
// a separate slot, like CatalogTriggerDoublers, so a game-package test
// can stub it without importing the catalog.
var CatalogTriggerSuppressors func(oracleID string) []TriggerSuppressor

// triggerOrigin says what kind of object a harvested ability belongs
// to. Only a suppressor reads it as such. A doubler reads one bit of
// it, FromSpell.
type triggerOrigin uint8

const (
	// triggerOfPermanent: a permanent's ability. It was found on the
	// battlefield, or it is a leaves-the-battlefield ability of the
	// permanent that just left (harvestLTB, the simultaneous-exit pass).
	// Evoke's sacrifice trigger is one too.
	triggerOfPermanent triggerOrigin = iota
	// triggerOfSpell: a spell's own "when you cast this spell"
	// (TriggeredAbility.FromStack).
	triggerOfSpell
	// triggerOfNonPermanent: an emblem's ability (CR 114.4), or a
	// card's ability that works from another zone (CR 113.6).
	triggerOfNonPermanent
)

// triggerSuppressedLocked reports whether a battlefield static stops
// this ability from triggering on this event. harvestMatchLocked calls
// it first, before the once-per-batch guard and before the doublers.
//
// It reads the same per-event candidate list as the doublers
// (scanTriggerModifiersLocked), so a game with no suppressor on the
// battlefield pays one nil check per candidate, and a game with no
// doubler or suppressor at all pays nothing beyond the scan the doublers
// already ran.
//
// Caller must hold g.mu in write mode.
func (g *Game) triggerSuppressedLocked(p *harvestPass, source Card, lki Characteristic, ability *TriggeredAbility, origin triggerOrigin) bool {
	if CatalogTriggerSuppressors == nil {
		return false
	}
	if !p.scanned {
		g.scanTriggerModifiersLocked(p)
	}
	if len(p.modifiers) == 0 {
		return false
	}
	// CR 603.10a: a leaves-the-battlefield ability looks back in time,
	// so a suppressor that left in the same event still counts. Any
	// other ability is judged on the board as it is after the event.
	lookBack := isBattlefieldExitEvent(p.ev)
	for _, candidate := range p.modifiers {
		if len(candidate.suppressors) == 0 || candidate.lki.AbilitiesRemoved {
			continue
		}
		if !lookBack && !candidate.live {
			continue
		}
		for _, s := range candidate.suppressors {
			if s.Suppresses == nil {
				continue
			}
			if s.ActiveWhen.IsGate() && !s.ActiveWhen.Active(candidate.card) {
				continue
			}
			q := TriggerSuppressionQuery{
				Event:             p.ev,
				Suppressor:        candidate.card,
				SuppressorLKI:     candidate.lki,
				Source:            source,
				SourceLKI:         lki,
				SourceIsPermanent: origin == triggerOfPermanent,
				Ability:           ability,
				Subject:           p.subject,
				SubjectLKI:        p.subjectLKI,
				HasSubject:        p.hasSubject,
			}
			if s.Suppresses(g, q) {
				return true
			}
		}
	}
	return false
}

// delayedTriggerSuppressedLocked asks the same question for an
// event-conditioned delayed trigger (CR 603.7). Torpor Orb's "don't
// cause abilities to trigger" makes no exception for delayed
// abilities. The concrete case on develop is an earthbent land: its
// "when it dies, return it" is a delayed trigger, and Hushbringer stops
// it.
//
// The source is the permanent that created the delayed trigger, if it
// is still on the battlefield (CR 603.7e). Otherwise the trigger counts
// as an ability of no permanent.
//
// Caller must hold g.mu in write mode.
func (g *Game) delayedTriggerSuppressedLocked(p *harvestPass, dt *DelayedTrigger) bool {
	if CatalogTriggerSuppressors == nil || dt == nil {
		return false
	}
	source, lki := g.triggerSourceLocked(dt.SourceCardID, dt.Controller)
	origin := triggerOfNonPermanent
	if dt.SourceCardID != uuid.Nil && findCardOnBattlefield(g, dt.SourceCardID) >= 0 {
		origin = triggerOfPermanent
	}
	return g.triggerSuppressedLocked(p, source, lki, nil, origin)
}
