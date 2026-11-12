package game

import (
	"errors"

	"github.com/google/uuid"
)

// enlist.go — CR 702.154, enlist (#5658).
//
//	702.154a Enlist represents a static ability and a triggered
//	         ability. Enlist means "As this creature attacks, you may
//	         tap up to one untapped creature you control that you didn't
//	         choose to attack with and that either has haste or has been
//	         under your control continuously since this turn began. When
//	         you do, this creature gets +X/+0 until end of turn, where X
//	         is the tapped creature's power."
//	702.154b Enlist's static ability represents an optional cost to
//	         attack (see rule 508.1g). Its triggered ability is linked
//	         to that static ability (see rule 607.2h).
//	702.154c A creature "enlists" another creature when you pay the
//	         cost of the creature's enlist ability by tapping the other
//	         creature. Note that it isn't possible for a creature to
//	         enlist itself.
//
// A KEYWORD TOKEN, like melee (melee.go): the deck importer stamps
// "enlist" from Scryfall's keywords array, so a printed enlist needs no
// catalog field, a creature that loses its abilities can't enlist, and
// the linked trigger is derived from the ability list
// (keywordTriggersFor).
//
// THE COST is exert's shape (exert.go, ADR 0130 §2), with a creature to
// name instead of a yes: both declaration verbs take the creature the
// attacker enlists (AttackDeclaration.Enlist), validate it before
// anything is paid or staged (ErrCantEnlist) and stage it on
// Card.EnlistOnAttack beside AttackingTarget. commitAttackDeclarationLocked
// pays every staged enlist as the declaration locks in (CR 508.1j): it
// taps the enlisted creature and emits EventEnlist, before the
// EventAttacks and in the same event batch, so the linked trigger and
// "whenever this creature enlists a creature" go on the stack with the
// attack triggers (CR 508.1m, 508.2). A staged attack cleared before
// the lock-in never tapped anything.
//
// WHO CAN BE ENLISTED (CR 702.154a, 702.154c), checked by
// enlistRefusalLocked at the verb and again as the cost is paid:
//
//   - a creature the attacker's controller controls, other than the
//     attacker itself;
//   - untapped;
//   - not chosen to attack with: not staged or declared as an attacker,
//     and not named as an attacker in the same declaration;
//   - with haste, or under that player's control continuously since the
//     turn began (HasSummoningSickness);
//   - not enlisted by another attacker of the same declaration: one
//     creature is tapped once.
//
// A creature staged as enlisted can't then be declared as an attacker
// in the same declaration (ErrEnlistedCantAttack): the player takes the
// declaration back (clear_combat, undo) to change their mind.
//
// A requirement never forces an enlist (CR 508.1d), and under a
// declaration another player chooses (declaration_control.go) the
// declarer can't enlist: the cost is the active player's.
//
// ONE INSTANCE. CR 702.154d lets several instances of enlist on one
// creature each enlist a creature. No card prints two or grants enlist,
// so the token is not cumulative and an attacker enlists at most one
// creature.
//
// THE TRIGGER. "When you do" watches EventEnlist for the source itself.
// The attacker and the enlisted creature are pinned by instance and
// object epoch as it triggers. X is read as the ability RESOLVES
// (CR 608.2h): the enlisted creature's power now, or as it last existed
// on the battlefield if it has left. A negative power adds nothing
// (CR 107.1b). The +X/+0 is a layer-7c scoped effect until end of turn
// on the attacker, which gets nothing if it is no longer the same
// object (CR 400.7).

// KeywordEnlist is CR 702.154's canonical token.
const KeywordEnlist = "enlist"

// enlistLabel is the stack label every enlist trigger carries, and the
// label of its +X/+0.
const enlistLabel = "Enlist — +X/+0 until end of turn, where X is the enlisted creature's power"

// enlistPumpKey is the body every enlist trigger names (ADR 0041 P9):
// an on-disk identity, never renamed or reused. Params.Objects is the
// attacker, then the creature it enlisted.
const enlistPumpKey = "enlist/pump"

// enlistPumpBody is the registered body's reference, built from the key
// and registered in init, as melee's is.
var enlistPumpBody = BodyRef{key: enlistPumpKey}

func init() {
	DelayedBody(enlistPumpKey, func(g *Game, item *StackItem, p EffectParams) error {
		if len(p.Objects) == 2 {
			g.resolveEnlistLocked(item, p.Objects[0], p.Objects[1])
		}
		return nil
	})
}

var (
	// ErrCantEnlist refuses an enlist the rules don't allow: the
	// attacker has no enlist or has already been declared this combat,
	// or the creature it names can't be enlisted (see the file comment).
	ErrCantEnlist = errors.New("game: that creature can't be enlisted by this attacker")
	// ErrEnlistedCantAttack refuses declaring as an attacker a creature
	// another attacker of this declaration has enlisted.
	ErrEnlistedCantAttack = errors.New("game: this creature has been enlisted, so it can't attack")
	// ErrDeclarerCantEnlist refuses an enlist sent by the declarer of a
	// chosen declaration: enlisting is a cost to attack, and the active
	// player chooses it.
	ErrDeclarerCantEnlist = errors.New("game: the creature's controller decides whether it enlists a creature as it attacks")
)

// enlistTrigger is the linked "When you do" of enlist (CR 702.154a,
// 607.2h). A package-level value, never mutated: keywordTriggersFor
// hands out copies of it.
var enlistTrigger = TriggeredAbility{
	Keyword: KeywordEnlist,
	Watches: []EventKind{EventEnlist},
	AppliesTo: func(ev Event, source *Card, _ Characteristic, _ *Game) bool {
		return ev.CardID != uuid.Nil && ev.CardID == source.InstanceID && ev.Target != uuid.Nil
	},
	Build: func(ev Event, source *Card, _ Characteristic, g *Game) *StackItem {
		self := ObjectRef{ID: source.InstanceID, Epoch: source.ObjectEpoch}
		enlisted := ObjectRef{ID: ev.Target, Epoch: g.objectEpochLocked(ev.Target)}
		return NewKeyedTriggeredItem(source, enlistLabel, enlistPumpBody, EffectParams{Objects: []ObjectRef{self, enlisted}})
	},
}

// resolveEnlistLocked is "this creature gets +X/+0 until end of turn,
// where X is the tapped creature's power", X read as the ability
// resolves (CR 608.2h).
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) resolveEnlistLocked(item *StackItem, self, enlisted ObjectRef) {
	if item == nil || self.ID == uuid.Nil {
		return
	}
	info, ok := g.PermanentForEffect(enlisted)
	if !ok || info.Power <= 0 {
		return
	}
	c := findBattlefieldCard(g, self.ID)
	if c == nil || c.ObjectEpoch != self.Epoch {
		return
	}
	g.RegisterScopedEffectForEffect(item.SourceCardID,
		[]AffectedObject{PinObject(c.InstanceID, c.EnteredBattlefieldAt)},
		[]Mod{ModifyPTMod(info.Power, 0)}, g.UntilEndOfTurnDuration(), enlistLabel)
}

// canEnlistLocked reports whether `attacker` may enlist a creature as
// it is declared now: it is on the battlefield with enlist, has not
// already been declared as an attacker this combat (the cost is chosen
// with the declaration, never later), and has no enlist staged.
//
// Caller must hold g.mu, with layers fresh.
func (g *Game) canEnlistLocked(attacker *Card) bool {
	if attacker == nil || findBattlefieldCard(g, attacker.InstanceID) == nil {
		return false
	}
	if g.announcedAttacks[attacker.InstanceID] {
		return false
	}
	return HasKeyword(attacker, KeywordEnlist)
}

// enlistedByLocked returns the attacker that has `id` staged as the
// creature it enlists, or nil.
//
// Caller must hold g.mu.
func (g *Game) enlistedByLocked(id uuid.UUID) *Card {
	if id == uuid.Nil || g.Battlefield == nil {
		return nil
	}
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].EnlistOnAttack == id {
			return &g.Battlefield.Cards[i]
		}
	}
	return nil
}

// enlistableLocked is CR 702.154a's description of the creature that
// may be tapped, for an attacker `controller` controls: an untapped
// creature that player controls, other than the attacker, that isn't
// attacking and either has haste or has been under their control
// continuously since the turn began.
//
// Caller must hold g.mu, with layers fresh.
func enlistableLocked(cand *Card, attackerID, controller uuid.UUID) bool {
	if cand == nil || cand.InstanceID == attackerID || cand.Controller != controller {
		return false
	}
	if !cand.IsCreature() || cand.Tapped || cand.AttackingTarget != uuid.Nil {
		return false
	}
	return !HasSummoningSickness(cand)
}

// EnlistCandidatesForEffect lists the creatures `attacker` may enlist
// as it is declared now, in battlefield order: the legal-move
// enumerator's read, through the function the verbs check (#544). Nil
// when the attacker can't enlist. A creature another attacker has
// already staged is left out.
//
// Caller must hold g.mu, with layers fresh.
func (g *Game) EnlistCandidatesForEffect(attacker *Card) []uuid.UUID {
	if !g.canEnlistLocked(attacker) || attacker.EnlistOnAttack != uuid.Nil {
		return nil
	}
	var out []uuid.UUID
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if !enlistableLocked(c, attacker.InstanceID, attacker.Controller) {
			continue
		}
		if g.enlistedByLocked(c.InstanceID) != nil {
			continue
		}
		out = append(out, c.InstanceID)
	}
	return out
}

// validateEnlistChoicesLocked checks one declaration's enlists before
// anything is paid or staged, so a refusal leaves the board as it was:
// every attacker that names a creature must be able to enlist it, no
// creature is enlisted twice, and no creature both attacks and is
// enlisted. A creature staged as enlisted earlier in this combat's
// declaration can't be declared as an attacker.
//
// An attacker re-declared naming the creature it has already staged is
// a re-point that keeps its choice.
//
// Caller must hold g.mu, with layers fresh.
func (g *Game) validateEnlistChoicesLocked(decls []AttackDeclaration) error {
	attacking := make(map[uuid.UUID]bool, len(decls))
	for _, d := range decls {
		attacking[d.Attacker] = true
	}
	taken := map[uuid.UUID]bool{}
	for _, d := range decls {
		if by := g.enlistedByLocked(d.Attacker); by != nil {
			return ErrEnlistedCantAttack
		}
		if d.Enlist == uuid.Nil {
			continue
		}
		attacker := findBattlefieldCard(g, d.Attacker)
		if attacker != nil && attacker.EnlistOnAttack == d.Enlist && !g.announcedAttacks[d.Attacker] {
			taken[d.Enlist] = true
			continue
		}
		if !g.canEnlistLocked(attacker) || attacking[d.Enlist] || taken[d.Enlist] {
			return ErrCantEnlist
		}
		if !enlistableLocked(findBattlefieldCard(g, d.Enlist), attacker.InstanceID, attacker.Controller) {
			return ErrCantEnlist
		}
		if by := g.enlistedByLocked(d.Enlist); by != nil && by.InstanceID != d.Attacker {
			return ErrCantEnlist
		}
		taken[d.Enlist] = true
	}
	return nil
}

// payEnlistLocked pays one staged enlist as the declaration locks in
// (CR 508.1j, 702.154c): the enlisted creature is tapped and EventEnlist
// is emitted. A creature that can no longer be enlisted (it left, was
// tapped or changed controller since the verb) is not tapped and
// nothing is enlisted: the cost can't be paid, so the optional cost
// isn't.
//
// Caller must hold g.mu in write mode.
func (g *Game) payEnlistLocked(attackerID, controller, enlistedID uuid.UUID) {
	cand := findBattlefieldCard(g, enlistedID)
	if !enlistableLocked(cand, attackerID, controller) {
		return
	}
	if err := g.setTapStateLocked(enlistedID, true); err != nil {
		return
	}
	g.EmitEvent(Event{
		Kind:   EventEnlist,
		Actor:  controller,
		CardID: attackerID,
		Source: attackerID,
		Target: enlistedID,
	})
}
