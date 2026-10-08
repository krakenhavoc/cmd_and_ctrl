package game

import "github.com/google/uuid"

// cant_attack_player.go — #2109, ADR 0063's 2026-10-08 amendment:
// "<player> can't attack you or permanents you control during their
// next turn" (The Second Doctor's How Civil of You), a CR 508.1c
// restriction on a PLAYER's attacks created by a resolving ability,
// with a CR 611.2 duration.
//
// # Where it sits
//
// An AttackTargetRestriction (ADR 0106 §2) is a clause on one
// CREATURE, written by a layer-6 static, and it dies with the creature.
// This restriction is on the attacker's PLAYER: it covers every
// creature that player controls now or later in the window, including
// one that has not entered yet. So it is a tenth payload on PlayerStatic
// (the slice of things a player has for a duration), stored on the
// restricted player and naming the protected one, and its reader sits
// beside the creature's own list in canAttackTargetWithLocked. Every
// consumer of that predicate follows for free: both declaration verbs,
// the CR 508.1d requirement search (so "attacks each combat if able"
// owes nothing against a player it may not attack), and the
// enumerator's per-attacker list (AttackTargetsForAttackerForEffect),
// which the ADR 0105 digest's attack_targets and the client's gate also
// read.
//
// # The window
//
// "During their next turn" is the restricted player's next turn from
// the moment the ability resolves: it neither starts early (an attacker
// whose own turn is under way when the grant is made is not restricted
// this turn) nor outlasts that turn. The duration is "until the end of
// their next turn" (UntilEndOfYourNextTurnDuration) and the grant
// additionally records the turn count it starts at, because the
// duration alone is already live now and an ability resolving in the
// restricted player's own turn would otherwise reach the turn in
// progress. CR 800.4m: a player who leaves the game takes it with them.
//
// # What it covers
//
// The protected player, and any permanent the protected player
// controls that can be attacked (planeswalkers, and battles they
// control). A battle's PROTECTOR is not the controller, and "permanents
// you control" does not name it, so a battle is covered by who controls
// it, never by who protects it.

// CantAttackGrant is a granted "can't attack <Protected> or permanents
// they control" statement. It lives on the RESTRICTED player's
// PlayerStatic. Plain data, like every other payload on that slice.
type CantAttackGrant struct {
	// Protected is whose attacks are refused: that player and the
	// permanents they control. uuid.Nil is the presence bit — the grant
	// says nothing.
	Protected uuid.UUID `json:"protected"`

	// FromTurnsBegun is the restricted player's seat-turn count at
	// which the restriction starts to apply: the grant's creation count
	// plus one, so it covers their NEXT turn and never the one in
	// progress. See the window above.
	FromTurnsBegun int `json:"fromTurnsBegun"`
}

// GrantCantAttackPlayerForEffect makes `attacker` unable to attack
// `protected` or the permanents `protected` controls during
// `attacker`'s next turn. The *ForEffect surface: caller must hold
// g.mu (write), which a resolution frame already does.
//
// A no-op when either player is missing or they are the same player (a
// player cannot attack themselves, so there is nothing to refuse).
// Several grants may live at once: two Second Doctors protect their two
// controllers.
func (g *Game) GrantCantAttackPlayerForEffect(attacker, protected uuid.UUID, label string, source uuid.UUID) {
	if attacker == uuid.Nil || protected == uuid.Nil || attacker == protected {
		return
	}
	p := g.playerByIDLocked(attacker)
	if p == nil || g.playerByIDLocked(protected) == nil {
		return
	}
	p.Statics = append(p.Statics, PlayerStatic{
		CantAttack: CantAttackGrant{
			Protected:      protected,
			FromTurnsBegun: g.turnsBegunForLocked(attacker) + 1,
		},
		Source:   source,
		Label:    label,
		Duration: g.UntilEndOfYourNextTurnDuration(attacker),
	})
}

// PlayerCantAttackError is the refusal a declaration verb returns when
// the attacking player has a live CantAttackGrant naming the
// defender. It wraps ErrIllegalAttackTarget, so every caller that
// tests for that keeps working, and its message is the toast's
// sentence.
type PlayerCantAttackError struct {
	// Attacker is the creature, AttackerName its name.
	Attacker     uuid.UUID
	AttackerName string
	// Protected is the player the attacker's controller may not attack,
	// ProtectedName their name, and Source the permanent that said so.
	Protected     uuid.UUID
	ProtectedName string
	Source        uuid.UUID
	SourceName    string
}

func (e *PlayerCantAttackError) Error() string { return "game: " + e.Sentence() }

// Unwrap makes errors.Is(err, ErrIllegalAttackTarget) hold.
func (e *PlayerCantAttackError) Unwrap() error { return ErrIllegalAttackTarget }

// Sentence is the refusal in a player's words.
func (e *PlayerCantAttackError) Sentence() string {
	s := e.AttackerName + " can't attack " + e.ProtectedName + " or permanents they control this turn"
	if e.SourceName != "" {
		s += " (" + e.SourceName + ")"
	}
	return s + "."
}

// playerCantAttackRefusalLocked is THE reader: nil when `attacker`'s
// controller has no live grant forbidding an attack on `target`, else
// the refusal naming the first that does.
//
// THE WINDOW AND THE DURATION ARE TESTED HERE, not left to the sweep,
// for the reason castBanForbidsLocked gives: the sweep is hygiene run
// at known moments and this read has to be right between them.
//
// Caller must hold g.mu (read or write) with fresh layers.
func (g *Game) playerCantAttackRefusalLocked(attacker *Card, target uuid.UUID) error {
	if attacker == nil {
		return nil
	}
	p := g.playerByIDLocked(attacker.Controller)
	if p == nil || len(p.Statics) == 0 {
		return nil
	}
	for _, s := range p.Statics {
		if s.CantAttack.Protected == uuid.Nil {
			continue
		}
		if g.turnsBegunForLocked(p.ID) < s.CantAttack.FromTurnsBegun {
			continue
		}
		if g.durationExpiredLocked(s.Duration, false) {
			continue
		}
		if !g.attackTargetIsOrBelongsToLocked(target, s.CantAttack.Protected) {
			continue
		}
		return &PlayerCantAttackError{
			Attacker:      attacker.InstanceID,
			AttackerName:  attacker.Effective().Name,
			Protected:     s.CantAttack.Protected,
			ProtectedName: g.playerNameLocked(s.CantAttack.Protected),
			Source:        s.Source,
			SourceName:    s.Label,
		}
	}
	return nil
}

// attackTargetIsOrBelongsToLocked reports whether `target` is the
// player `who`, or a permanent `who` controls. A battle counts by its
// controller, not its protector (see the file comment).
//
// Caller must hold g.mu.
func (g *Game) attackTargetIsOrBelongsToLocked(target, who uuid.UUID) bool {
	if target == who {
		return true
	}
	c := findBattlefieldCard(g, target)
	return c != nil && c.Controller == who
}

// playerHasCantAttackGrantLocked is the cheap early-out for the
// enumerator's per-attacker list: does this player carry any
// CantAttackGrant at all (live or not)? Caller must hold g.mu.
func (g *Game) playerHasCantAttackGrantLocked(player uuid.UUID) bool {
	p := g.playerByIDLocked(player)
	if p == nil {
		return false
	}
	for _, s := range p.Statics {
		if s.CantAttack.Protected != uuid.Nil {
			return true
		}
	}
	return false
}
