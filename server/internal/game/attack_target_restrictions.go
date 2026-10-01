package game

import (
	"github.com/google/uuid"
)

// attack_target_restrictions.go is ADR 0106 §2 (#1794): a CR 508.1c
// restriction on WHAT a creature may attack, keyed on the creature
// rather than on its controller.
//
// Xantcha, Sleeper Agent "can't attack its owner or planeswalkers its
// owner controls"; Alexios, Deimos of Kosmos "can't attack its owner";
// Elrond of the White Council's stolen creatures gain "This creature
// can't attack its owner." Each names a defender by its relation to
// the ATTACKER, so the old target check — canAttackTargetLocked, which
// is CR 506.2 for one seat and takes only the controller — cannot say
// it. canAttackTargetWithLocked is that check plus this list, and
// every caller that judges a DECLARATION moves to it:
//
//   - both declaration verbs (DeclareAttackerWith, DeclareAttackersWith),
//   - the CR 508.1d search (attackRequirementCandidatesLocked), so the
//     maximum is counted over restriction-legal targets only: with only
//     its owner to attack, Xantcha does not attack, and nothing is owed,
//   - the legal-move enumerator, through AttackTargetsForAttackerForEffect,
//     and the ADR 0105 digest built from its moves.
//
// Two callers deliberately keep the controller-only check, because the
// rules exempt them from "requirements or restrictions that apply to
// the declaration of attackers":
//
//   - CR 508.7b, reselecting what a creature attacks (attack_reselect.go),
//   - CR 508.4c, a creature put onto the battlefield attacking.
//
// # Why data on the characteristic and not a Restriction bit
//
// A bit could not say "and planeswalkers its owner controls", which
// Xantcha prints and Alexios does not, and a bit names no source for
// the refusal sentence or the card's chip. So it is Restrictions' and
// AttackRequirements' twin: written by an ordinary layer static, only
// ever appended to, never cleared by a layer-6 ability removal. A
// creature's OWN restriction still goes with its abilities, because the
// catalog static that writes it is not applied once CatalogAbilityKey
// answers empty (CR 613.1f).
//
// The owner is read LIVE off the attacking creature (CR 108.3, 111.2),
// never captured: a Clone copying Xantcha is owned by the Clone's
// owner, and that is who it may not attack.
//
// Battles are never refused by the owner clauses. "Planeswalkers its
// owner controls" says nothing about a battle the owner protects, so
// attacking one is legal, and attacking it obeys "attacks each combat if
// able".
//
// # "Can't attack unless defending player controls an Island"
//
// ADR 0107 §2 (#1879) is the third clause, DefenderMustControl. CR 508.5
// makes "defending player" a fact about each TARGET: the player attacked,
// the controller of the planeswalker attacked, or the protector of the
// battle attacked. CR 508.5a makes it one specific player, worked out
// separately for each creature. So in Commander, Sea Serpent may attack an
// opponent who controls an Island, a planeswalker that opponent controls
// and a battle that opponent protects, and nothing else. The clause is
// per target, which is this file's shape, and
// defendingPlayerForAttackLocked already answers CR 508.5 for every
// target kind.
//
// The board is read live, at declaration, which is the only time CR 508.1c
// asks. An Island that leaves after attackers are declared changes
// nothing (CR 506.4a).

// AttackTargetRestriction is one CR 508.1c restriction on what one
// creature may attack. Pure data, so the Characteristic that carries it
// stays copyable and the snapshot's lastKnownBattlefield round-trips it.
type AttackTargetRestriction struct {
	// Source is the permanent whose text imposes it.
	Source uuid.UUID
	// SourceName is that source's name, for the refusal sentence and
	// the card's chip. Captured when the restriction is written.
	SourceName string
	// NotOwner forbids attacking the creature's owner.
	NotOwner bool
	// NotOwnersPlaneswalkers forbids attacking a planeswalker the
	// creature's owner controls.
	NotOwnersPlaneswalkers bool
	// DefenderMustControl, when set, forbids attacking any target whose
	// defending player (CR 508.5) controls no permanent matching ANY of
	// these queries: "can't attack unless defending player controls an
	// Island" (ADR 0107 §2, #1879). Plain data, so the Characteristic
	// stays snapshot-safe.
	DefenderMustControl []PermanentQuery `json:",omitempty"`
}

// Equal reports whether r and o are the same restriction. A method rather
// than ==, because DefenderMustControl is a slice.
func (r AttackTargetRestriction) Equal(o AttackTargetRestriction) bool {
	return r.Source == o.Source && r.SourceName == o.SourceName &&
		r.NotOwner == o.NotOwner && r.NotOwnersPlaneswalkers == o.NotOwnersPlaneswalkers &&
		samePermanentQueries(r.DefenderMustControl, o.DefenderMustControl)
}

// AttackTargetRestrictionError is the refusal a declaration verb
// returns when a creature is pointed at a target one of its
// AttackTargetRestrictions forbids. It wraps ErrIllegalAttackTarget, so
// every caller that already tests for that keeps working, and its
// message is the sentence the client's toast shows.
type AttackTargetRestrictionError struct {
	// Attacker is the creature, and AttackerName its name.
	Attacker     uuid.UUID
	AttackerName string
	// Restriction is the one that refused it.
	Restriction AttackTargetRestriction
	// For a DefenderMustControl refusal: TargetName is what was
	// attacked (a player's or a permanent's name), and DefenderName the
	// defending player who controls no matching permanent.
	TargetName   string
	DefenderName string
}

func (e *AttackTargetRestrictionError) Error() string {
	return "game: " + e.Sentence()
}

// Unwrap makes errors.Is(err, ErrIllegalAttackTarget) hold.
func (e *AttackTargetRestrictionError) Unwrap() error { return ErrIllegalAttackTarget }

// Sentence is the refusal in a player's words, with the source named
// when it is not the creature itself ("Grizzly Bears can't attack its
// owner (Elrond of the White Council).").
func (e *AttackTargetRestrictionError) Sentence() string {
	if e.DefenderName != "" {
		// "Sea Serpent can't attack Bob: Bob controls no Island."
		s := e.AttackerName + " can't attack " + e.TargetName + ": " +
			e.DefenderName + " controls no " + PermanentQueriesNoun(e.Restriction.DefenderMustControl)
		if e.Restriction.SourceName != "" && e.Restriction.Source != e.Attacker {
			s += " (" + e.Restriction.SourceName + ")"
		}
		return s + "."
	}
	what := "its owner"
	switch {
	case e.Restriction.NotOwner && e.Restriction.NotOwnersPlaneswalkers:
		what = "its owner or planeswalkers its owner controls"
	case e.Restriction.NotOwnersPlaneswalkers:
		what = "planeswalkers its owner controls"
	}
	s := e.AttackerName + " can't attack " + what
	if e.Restriction.SourceName != "" && e.Restriction.Source != e.Attacker {
		s += " (" + e.Restriction.SourceName + ")"
	}
	return s + "."
}

// refuses reports whether r forbids `attacker` attacking `target`, and,
// for a DefenderMustControl refusal, the defending player who controls
// nothing it asks for.
//
// Caller must hold g.mu with fresh layers.
func (r AttackTargetRestriction) refuses(g *Game, attacker *Card, target uuid.UUID) (bool, uuid.UUID) {
	if r.ownerRefuses(g, attacker, target) {
		return true, uuid.Nil
	}
	if len(r.DefenderMustControl) > 0 {
		// CR 508.5: the player attacked, a planeswalker's controller,
		// or a battle's protector.
		defender := g.defendingPlayerForAttackLocked(target)
		if defender != uuid.Nil && !g.controlsMatchingLocked(defender, r.DefenderMustControl) {
			return true, defender
		}
	}
	return false, uuid.Nil
}

// ownerRefuses is the two owner clauses (ADR 0106 §2).
//
// Caller must hold g.mu with fresh layers.
func (r AttackTargetRestriction) ownerRefuses(g *Game, attacker *Card, target uuid.UUID) bool {
	owner := attacker.Owner
	if owner == uuid.Nil || (!r.NotOwner && !r.NotOwnersPlaneswalkers) {
		return false
	}
	switch g.classifyAttackTargetLocked(target) {
	case AttackTargetPlayer:
		return r.NotOwner && target == owner
	case AttackTargetPlaneswalker:
		if !r.NotOwnersPlaneswalkers {
			return false
		}
		pw := findBattlefieldCard(g, target)
		return pw != nil && pw.Controller == owner
	}
	return false
}

// attackTargetRestrictionRefusalLocked is the restriction half of
// canAttackTargetWithLocked: nil when none of `attacker`'s
// AttackTargetRestrictions forbids `target`, else the refusal naming
// the first that does.
//
// Caller must hold g.mu with fresh layers.
func (g *Game) attackTargetRestrictionRefusalLocked(attacker *Card, target uuid.UUID) error {
	if attacker == nil {
		return nil
	}
	for _, r := range attacker.Effective().AttackTargetRestrictions {
		refused, defender := r.refuses(g, attacker, target)
		if !refused {
			continue
		}
		err := &AttackTargetRestrictionError{
			Attacker:     attacker.InstanceID,
			AttackerName: attacker.Effective().Name,
			Restriction:  r,
		}
		if defender != uuid.Nil {
			err.DefenderName = g.playerNameLocked(defender)
			err.TargetName = err.DefenderName
			if c := findBattlefieldCard(g, target); c != nil {
				err.TargetName = c.Effective().Name
			}
		}
		return err
	}
	return nil
}

// playerNameLocked is a seat's name for a sentence, or "that player".
//
// Caller must hold g.mu.
func (g *Game) playerNameLocked(id uuid.UUID) string {
	if p := g.playerByIDLocked(id); p != nil && p.Name != "" {
		return p.Name
	}
	return "that player"
}

// DefenderRefusal is one opponent a creature can't attack right now under
// a DefenderMustControl restriction, for the card's chip (ADR 0107 §2
// decision 3): "can't attack Bob: Bob controls no Island".
type DefenderRefusal struct {
	// Player is the opponent who controls nothing the restriction asks
	// for. The refusal covers their planeswalkers and the battles they
	// protect too (CR 508.5).
	Player uuid.UUID
	// Restriction is the one that refuses them.
	Restriction AttackTargetRestriction
}

// DefenderRefusalsForEffect lists, for each DefenderMustControl
// restriction on `c`, every live opponent of its controller who controls
// no matching permanent right now. Nil for nearly every card.
//
// Read-only. Caller must hold g.mu with fresh layers.
func (g *Game) DefenderRefusalsForEffect(c *Card) []DefenderRefusal {
	if c == nil {
		return nil
	}
	var out []DefenderRefusal
	for _, r := range c.Effective().AttackTargetRestrictions {
		if len(r.DefenderMustControl) == 0 {
			continue
		}
		for _, p := range g.Seats {
			if p == nil || p.Eliminated || p.ID == c.Controller {
				continue
			}
			if !g.controlsMatchingLocked(p.ID, r.DefenderMustControl) {
				out = append(out, DefenderRefusal{Player: p.ID, Restriction: r})
			}
		}
	}
	return out
}

// canAttackTargetWithLocked reports whether `attacker` may be DECLARED
// attacking `target`: CR 506.2 and 310.9b for its controller
// (canAttackTargetLocked), then its own CR 508.1c target restrictions.
//
// The declaration-time check. CR 508.7b's reselection and CR 508.4c's
// "put onto the battlefield attacking" are exempt from restrictions on
// the declaration and keep canAttackTargetLocked.
//
// Caller must hold g.mu with fresh layers.
func (g *Game) canAttackTargetWithLocked(attacker *Card, target uuid.UUID) error {
	if attacker == nil {
		return ErrIllegalAttackTarget
	}
	if err := g.canAttackTargetLocked(attacker.Controller, target); err != nil {
		return err
	}
	return g.attackTargetRestrictionRefusalLocked(attacker, target)
}

// AttackTargetsForAttackerForEffect lists every target `attacker` may
// be declared attacking right now: AttackTargetsForEffect for its
// controller, less what its own target restrictions forbid. The
// enumerator's per-attacker list (ADR 0106 §2 decision 2).
//
// AttackTargetsForEffect stays per seat for the view's per-target
// prices and limits; legality per creature comes from here.
//
// Read-only. Caller must hold g.mu with fresh layers.
func (g *Game) AttackTargetsForAttackerForEffect(attacker *Card) []AttackTargetRef {
	if attacker == nil {
		return nil
	}
	all := g.AttackTargetsForEffect(attacker.Controller)
	if len(attacker.Effective().AttackTargetRestrictions) == 0 {
		return all
	}
	out := all[:0:0]
	for _, t := range all {
		if g.attackTargetRestrictionRefusalLocked(attacker, t.ID) == nil {
			out = append(out, t)
		}
	}
	return out
}
