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
//
// # This turn, and narrower scopes (#2719)
//
// ADR 0063's 2026-10-10 amendment adds a second window and narrower
// things a grant may cover. "They can't attack you or planeswalkers you
// control this turn" (Sandswirl Wanderglyph) is
// GrantCantAttackPlayerThisTurnForEffect: the floor is the restricted
// player's CURRENT seat-turn count and the duration is "until end of
// turn", so it reaches the turn in progress and no other. The
// CantAttackScope on the grant narrows the protected half: a
// planeswalker-only grant leaves a battle the protected player controls
// open (the Wanderglyph's ruling), a player-only one leaves every
// permanent open ("creatures they control can't attack you this turn",
// Web of Inertia), and a player-exempt one with a subtype covers only
// those planeswalkers ("can't attack Jaces you control this turn",
// Jace, Multiverse Architect). The zero scope is The Second Doctor's
// "you or permanents you control", so every grant made before it reads
// as it did.

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

	// Scope narrows what of Protected's the grant covers. The zero
	// scope is the player and every permanent they control.
	//
	// No `omitzero`, for PlayerStatic.Timing's reason (#1492).
	Scope CantAttackScope `json:"scope"`
}

// CantAttackScope narrows a CantAttackGrant (#2719). Each field only
// ever removes something from the default "the player and permanents
// they control", so the zero scope is the widest grant.
type CantAttackScope struct {
	// PlayerOnly covers the protected player and none of their
	// permanents: "creatures they control can't attack you".
	PlayerOnly bool `json:"playerOnly,omitempty"`

	// PlayerExempt leaves the protected player open and covers only
	// their permanents: "can't attack Jaces you control".
	PlayerExempt bool `json:"playerExempt,omitempty"`

	// PlaneswalkersOnly covers, of their permanents, planeswalkers
	// alone: "you or planeswalkers you control". A battle they control
	// is not covered.
	PlaneswalkersOnly bool `json:"planeswalkersOnly,omitempty"`

	// Subtype, when set, narrows the covered permanents to that
	// subtype: "Jaces you control" is PlaneswalkersOnly with "Jace".
	Subtype string `json:"subtype,omitempty"`
}

// covers reports whether the attack target `target` is something this
// scope protects, given that `who` is the protected player. A battle
// counts by its controller, not its protector (see the file comment).
//
// Caller must hold g.mu with fresh layers.
func (s CantAttackScope) covers(g *Game, target, who uuid.UUID) bool {
	if target == who {
		return !s.PlayerExempt
	}
	if s.PlayerOnly {
		return false
	}
	c := findBattlefieldCard(g, target)
	if c == nil || c.Controller != who {
		return false
	}
	if s.PlaneswalkersOnly && !c.IsPlaneswalker() {
		return false
	}
	return s.Subtype == "" || c.HasSubtype(s.Subtype)
}

// protectedPhrase is what the scope covers, in a refusal's words:
// "Alice or permanents they control", "Jaces Alice controls".
func (s CantAttackScope) protectedPhrase(name string) string {
	kind := "permanents"
	if s.PlaneswalkersOnly {
		kind = "planeswalkers"
	}
	if s.Subtype != "" {
		kind = s.Subtype + "s"
	}
	switch {
	case s.PlayerOnly:
		return name
	case s.PlayerExempt:
		return kind + " " + name + " controls"
	default:
		return name + " or " + kind + " they control"
	}
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
	g.grantCantAttackLocked(attacker, protected, CantAttackScope{}, g.turnsBegunForLocked(attacker)+1,
		g.UntilEndOfYourNextTurnDuration(attacker), label, source)
}

// GrantCantAttackPlayerThisTurnForEffect makes `attacker` unable to
// attack what `scope` covers of `protected`'s for the rest of the
// current turn (#2719): Sandswirl Wanderglyph's "they can't attack you
// or planeswalkers you control this turn", Web of Inertia's "creatures
// they control can't attack you this turn". The window opens now (the
// floor is the attacker's current seat-turn count) and closes as this
// turn ends, so on any turn but the attacker's own it restricts
// nothing. Same no-ops and caller contract as
// GrantCantAttackPlayerForEffect.
func (g *Game) GrantCantAttackPlayerThisTurnForEffect(attacker, protected uuid.UUID, scope CantAttackScope, label string, source uuid.UUID) {
	g.grantCantAttackLocked(attacker, protected, scope, g.turnsBegunForLocked(attacker),
		g.UntilEndOfTurnDuration(), label, source)
}

// grantCantAttackLocked is both writers' one body. Caller must hold
// g.mu (write).
func (g *Game) grantCantAttackLocked(attacker, protected uuid.UUID, scope CantAttackScope, from int, d Duration, label string, source uuid.UUID) {
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
			FromTurnsBegun: from,
			Scope:          scope,
		},
		Source:   source,
		Label:    label,
		Duration: d,
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
	// Scope is what of the protected player's the grant covers, for
	// the sentence.
	Scope CantAttackScope
}

func (e *PlayerCantAttackError) Error() string { return "game: " + e.Sentence() }

// Unwrap makes errors.Is(err, ErrIllegalAttackTarget) hold.
func (e *PlayerCantAttackError) Unwrap() error { return ErrIllegalAttackTarget }

// Sentence is the refusal in a player's words.
func (e *PlayerCantAttackError) Sentence() string {
	s := e.AttackerName + " can't attack " + e.Scope.protectedPhrase(e.ProtectedName) + " this turn"
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
		if !s.CantAttack.Scope.covers(g, target, s.CantAttack.Protected) {
			continue
		}
		return &PlayerCantAttackError{
			Attacker:      attacker.InstanceID,
			AttackerName:  attacker.Effective().Name,
			Protected:     s.CantAttack.Protected,
			ProtectedName: g.playerNameLocked(s.CantAttack.Protected),
			Source:        s.Source,
			SourceName:    s.Label,
			Scope:         s.CantAttack.Scope,
		}
	}
	return nil
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
