package game

import (
	"errors"

	"github.com/google/uuid"
)

// exert.go — exert (CR 701.43, ADR 0130).
//
// One primitive, exertLocked, is the only way a permanent is exerted.
// It keeps the permanent from untapping during the EXERTING player's
// next untap step by reusing ADR 0058's next-untap marker
// (SkipNextUntapForEffect), records the exert on the turn's tally
// (TurnTally.Exerts) and emits EventExert, which "when you do" and
// "whenever you exert a creature" watch.
//
// "You may exert this creature as it attacks" is an optional cost to
// attack (CR 701.43d, 508.1g). The catalog declares it on the card
// (CardDef.ExertOnAttack). The two declaration verbs take the choice
// per attacker and stage it on Card.ExertOnAttack beside
// AttackingTarget. commitAttackDeclarationLocked pays every staged
// exert as the declaration locks in, before it announces the attacks
// and in the same event batch (CR 508.1j before 508.1k and 508.1m;
// ADR 0130 §2, owner decision 2), so the triggers the exerts cause go
// on the stack with the attack triggers, before blockers.
//
// The marker's duplicate rule is load-bearing here: a second exert
// before the exerter's next untap step adds no second marker, so every
// exert expires at that one step (CR 701.43b). Whoever makes the plain
// marker count (#2029, Telekinesis's "next two untap steps") must not
// change that; see ADR 0130 §1.

// ExertOnAttack is the static ability "You may exert this creature as
// it attacks" (CR 701.43d), declared on a catalog card.
type ExertOnAttack struct {
	// Unless is a condition under which the creature may NOT be
	// exerted as it attacks: Combat Celebrant's "If this creature
	// hasn't been exerted this turn" is Unless = "it has been". Nil
	// for every other card. Read with g.mu held.
	Unless func(g *Game, source *Card) bool
}

// CatalogExertOnAttack returns the exert-as-it-attacks ability a
// catalog entry declares, or nil. Wired to the card catalog in
// carddef.go from CardDef.ExertOnAttack; nil in a test that builds
// cards by hand. Keyed by CatalogAbilityKey, so a creature that has
// lost its abilities (CR 613.1f, Humility) can't be exerted.
var CatalogExertOnAttack func(key string) *ExertOnAttack

// ErrCantExert refuses `exert: true` on an attacker that can't be
// exerted as it attacks right now: it has no such ability, its
// condition fails (Combat Celebrant exerted earlier this turn), or it
// has already been declared as an attacker this combat (ruling: you
// can't exert later in combat).
var ErrCantExert = errors.New("game: this creature can't be exerted as it attacks")

// ExertRecord is one exert this turn (ADR 0130 §1): the permanent, its
// object epoch as it was exerted (CR 400.7: a permanent that left and
// came back is a new object), the player who exerted it, and the phase
// it happened in.
type ExertRecord struct {
	Object  uuid.UUID `json:"object"`
	Epoch   int       `json:"epoch,omitempty"`
	Player  uuid.UUID `json:"player,omitempty"`
	PhaseID int       `json:"phaseId,omitempty"`
}

// exertOnAttackOfLocked returns the exert-as-it-attacks ability `c`
// has right now, read off its effective abilities, or nil.
//
// Caller must hold g.mu.
func exertOnAttackOfLocked(c *Card) *ExertOnAttack {
	if c == nil || CatalogExertOnAttack == nil {
		return nil
	}
	key := catalogAbilityKeyOf(c)
	if key == "" {
		return nil
	}
	return CatalogExertOnAttack(key)
}

// canExertAsItAttacksLocked reports whether `c` may be exerted as it
// is declared as an attacker now: it is on the battlefield, has the
// ability, its Unless condition is false, and it has not already been
// declared as an attacker this combat (an exert is chosen with the
// declaration, never later).
//
// Caller must hold g.mu, with layers fresh.
func (g *Game) canExertAsItAttacksLocked(c *Card) bool {
	if c == nil || findBattlefieldCard(g, c.InstanceID) == nil {
		return false
	}
	if g.announcedAttacks[c.InstanceID] {
		return false
	}
	ab := exertOnAttackOfLocked(c)
	if ab == nil {
		return false
	}
	return ab.Unless == nil || !ab.Unless(g, c)
}

// CanExertAsItAttacksForEffect is canExertAsItAttacksLocked for the
// legal-move enumerator: whether an exert move is offered beside the
// plain attack. Caller must hold g.mu, with layers fresh (the
// enumerator's posture).
func (g *Game) CanExertAsItAttacksForEffect(c *Card) bool {
	return g.canExertAsItAttacksLocked(c)
}

// exertLocked exerts the battlefield permanent `cardID` for `player`
// (CR 701.43a): one next-untap marker keyed to `player`'s untap step,
// one ExertRecord on the turn's tally and one EventExert. `target` is
// the attack target when the permanent was exerted as it attacked, and
// uuid.Nil when it was exerted to pay a cost.
//
// A permanent that is not on the battlefield can't be exerted
// (CR 701.43c): nothing happens and false comes back.
//
// Caller must hold g.mu in write mode.
func (g *Game) exertLocked(cardID, player, target uuid.UUID) bool {
	c := findBattlefieldCard(g, cardID)
	if c == nil || player == uuid.Nil {
		return false
	}
	_ = g.SkipNextUntapForEffect(cardID, player)
	rec := ExertRecord{
		Object:  cardID,
		Epoch:   c.ObjectEpoch,
		Player:  player,
		PhaseID: g.Turn.PhaseID,
	}
	// A fresh slice every time, never an append into shared capacity:
	// the undo clone holds this backing array too (#1238's reason).
	exerts := make([]ExertRecord, len(g.TurnTally.Exerts), len(g.TurnTally.Exerts)+1)
	copy(exerts, g.TurnTally.Exerts)
	g.TurnTally.Exerts = append(exerts, rec)
	g.EmitEvent(Event{
		Kind:   EventExert,
		Actor:  player,
		CardID: cardID,
		Source: cardID,
		Target: target,
	})
	return true
}

// ExertedThisTurn reports whether the OBJECT `cardID` names now has
// been exerted this turn, by anyone (Combat Celebrant's "if this
// creature hasn't been exerted this turn"). Per object: a permanent
// that left the battlefield and came back is a new object that hasn't
// (CR 400.7).
//
// Caller must hold g.mu.
func (g *Game) ExertedThisTurn(cardID uuid.UUID) bool {
	if cardID == uuid.Nil || len(g.TurnTally.Exerts) == 0 {
		return false
	}
	epoch := g.objectEpochLocked(cardID)
	for _, e := range g.TurnTally.Exerts {
		if e.Object == cardID && e.Epoch == epoch {
			return true
		}
	}
	return false
}

// validateExertChoicesLocked checks every declaration that asks to
// exert: each one must name a creature that may be exerted as it
// attacks now. Called by both verbs before anything is paid or staged,
// so a refusal leaves the board exactly as it was (ADR 0130 §2: an
// exert flag the creature can't honour is a client bug, and dropping
// it silently would attack without the cost the player chose).
//
// Caller must hold g.mu, with layers fresh.
func (g *Game) validateExertChoicesLocked(decls []AttackDeclaration) error {
	for _, d := range decls {
		if !d.Exert {
			continue
		}
		if !g.canExertAsItAttacksLocked(findBattlefieldCard(g, d.Attacker)) {
			return ErrCantExert
		}
	}
	return nil
}
