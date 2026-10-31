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
//
// # The other conditions on the defending player
//
// The rest of #1879 is four more clauses of the same shape, each a fact
// about the target's defending player read live at declaration:
//
//   - DefenderMustBePoisoned, "unless defending player is poisoned"
//     (Chained Throatseeker): one or more poison counters (CR 122.1f);
//   - DefenderMustBeMonarch, "unless defending player is the monarch"
//     (Crown-Hunter Hireling, CR 725.1);
//   - DefenderGraveyardAtLeast, "unless defending player has seven or
//     more cards in their graveyard" (Vantress Gargoyle);
//   - ControllerMustControlMore, "unless you control more creatures than
//     defending player" (Goblin Goon, Mogg Toady; Monstrous Hound's
//     lands). The "you" is the attacking creature's controller: every
//     printed one is the creature's own ability;
//   - ControllerMustHaveCitysBlessing, "unless you have the city's
//     blessing" (Wayward Swordtooth, #2696, CR 702.131c). A fact about
//     the attacking creature's controller alone, so it refuses every
//     target while it is unmet.
//
// Every clause a restriction sets must hold. A creature that prints two
// conditions has two restrictions, and both must hold too.
//
// A restriction the creature imposes on ITSELF (Source is the creature)
// is always one of its own abilities: a printed one, or one a resolved
// effect gave it, as Veiled Serpent's trigger does. So a later "loses
// all abilities" takes it (CR 613.1f, 613.7), which the layer pass's
// removal does with dropSelfAttackTargetRestrictions. A restriction
// another permanent imposes is that permanent's ability and stays.

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
	// DefenderMustBePoisoned forbids attacking any target whose
	// defending player has no poison counters: "can't attack unless
	// defending player is poisoned" (CR 122.1f).
	DefenderMustBePoisoned bool `json:",omitempty"`
	// DefenderMustBeMonarch forbids attacking any target whose defending
	// player is not the monarch (CR 725.1).
	DefenderMustBeMonarch bool `json:",omitempty"`
	// DefenderGraveyardAtLeast, when above zero, forbids attacking any
	// target whose defending player has fewer cards than this in their
	// graveyard.
	DefenderGraveyardAtLeast int `json:",omitempty"`
	// ControllerMustControlMore, when set, forbids attacking any target
	// whose defending player controls at least as many permanents
	// matching any of these queries as the attacking creature's
	// controller does: "can't attack unless you control more creatures
	// than defending player".
	ControllerMustControlMore []PermanentQuery `json:",omitempty"`
	// ControllerMustHaveCitysBlessing forbids attacking anything while
	// the attacking creature's controller lacks the city's blessing:
	// "can't attack unless you have the city's blessing" (Wayward
	// Swordtooth, #2696, CR 702.131c). Read live at declaration, like
	// every clause here; the blessing is never lost once earned.
	ControllerMustHaveCitysBlessing bool `json:",omitempty"`
	// NotAlreadyAttackedThisTurn forbids attacking a player this object
	// has already attacked this turn: "can't attack a player it has
	// already attacked this turn" (Bloodthirster, #2171, CR 508.1c). It
	// reads TurnTally.Attacks for the attacker's current object epoch, so
	// a creature that left and came back is a new object (CR 400.7) with
	// no history. Players only: a planeswalker or battle is not "a
	// player", and an earlier attack on one is not an attack on its
	// controller.
	NotAlreadyAttackedThisTurn bool `json:",omitempty"`
	// NotPlayer, when set, forbids attacking that one player and the
	// planeswalkers they control: "can't attack you or planeswalkers you
	// control" given by a resolved effect (Promise of Loyalty, #3400,
	// CR 508.1c). Unlike the owner clauses it names a FIXED player, the
	// "you" of the resolving spell captured when it resolved (CR 611.2c),
	// not a relation read live off the attacker. A battle that player
	// protects is not covered, because the text does not say it.
	NotPlayer uuid.UUID `json:",omitempty"`
	// NotPlayerOnly narrows NotPlayer to the player alone: the planeswalkers
	// they control stay open to attack. "Creatures with flying can't attack
	// you" (Storm, Windrider, #4601) says "you", not "or planeswalkers you
	// control", so a flier may still attack one.
	NotPlayerOnly bool `json:",omitempty"`
	// OnlyAttackersMatching, when set, makes the whole restriction bind
	// only an attacker that matches ANY of these queries, judged on its
	// effective characteristics at the moment of declaration (CR 508.1c):
	// the flying in "creatures with flying can't attack you" is read live,
	// so a creature that loses or gains flying mid-turn is judged as it
	// stands when attackers are declared. The static writes the
	// restriction onto every creature and this filter does the choosing,
	// rather than the static's AppliesTo, because AppliesTo runs inside the
	// layer pass where a keyword another effect grants may not have landed
	// yet (#4601).
	OnlyAttackersMatching []PermanentQuery `json:",omitempty"`
	// PlayerMustHaveActedLastTurn forbids attacking a PLAYER who neither
	// cast a spell nor put a nontoken permanent onto the battlefield
	// during their last turn: Arboria's "Creatures can't attack a player
	// unless that player cast a spell or put a nontoken permanent onto
	// the battlefield during their last turn" (#1997, CR 508.1c). Read
	// live at declaration off Player.LastTurnCastOrPut. Players only: a
	// planeswalker or a battle may still be attacked (the 2022 ruling),
	// so it is not a defending-player clause.
	PlayerMustHaveActedLastTurn bool `json:",omitempty"`
}

// Equal reports whether r and o are the same restriction. A method rather
// than ==, because DefenderMustControl is a slice.
func (r AttackTargetRestriction) Equal(o AttackTargetRestriction) bool {
	return r.Source == o.Source && r.SourceName == o.SourceName &&
		r.NotOwner == o.NotOwner && r.NotOwnersPlaneswalkers == o.NotOwnersPlaneswalkers &&
		samePermanentQueries(r.DefenderMustControl, o.DefenderMustControl) &&
		r.DefenderMustBePoisoned == o.DefenderMustBePoisoned &&
		r.DefenderMustBeMonarch == o.DefenderMustBeMonarch &&
		r.DefenderGraveyardAtLeast == o.DefenderGraveyardAtLeast &&
		samePermanentQueries(r.ControllerMustControlMore, o.ControllerMustControlMore) &&
		r.NotAlreadyAttackedThisTurn == o.NotAlreadyAttackedThisTurn && r.NotPlayer == o.NotPlayer &&
		r.NotPlayerOnly == o.NotPlayerOnly &&
		samePermanentQueries(r.OnlyAttackersMatching, o.OnlyAttackersMatching) &&
		r.ControllerMustHaveCitysBlessing == o.ControllerMustHaveCitysBlessing &&
		r.PlayerMustHaveActedLastTurn == o.PlayerMustHaveActedLastTurn
}

// inactiveLastTurnRefuses is the PlayerMustHaveActedLastTurn clause:
// `target` is a player who cast no spell and put no nontoken permanent
// onto the battlefield during their last turn.
//
// Caller must hold g.mu.
func (r AttackTargetRestriction) inactiveLastTurnRefuses(g *Game, target uuid.UUID) bool {
	return r.PlayerMustHaveActedLastTurn && g.classifyAttackTargetLocked(target) == AttackTargetPlayer &&
		!g.CastOrPutDuringLastTurnForEffect(target)
}

// inactiveLastTurnReason is that clause's reason, in a player's words.
func inactiveLastTurnReason(name string) string {
	return name + " didn't cast a spell or put a nontoken permanent onto the battlefield during their last turn"
}

// asksOfTheDefender reports whether r has any clause about the target's
// defending player (CR 508.5), as opposed to only the owner clauses.
func (r AttackTargetRestriction) asksOfTheDefender() bool {
	return len(r.DefenderMustControl) > 0 || r.DefenderMustBePoisoned || r.DefenderMustBeMonarch ||
		r.DefenderGraveyardAtLeast > 0 || len(r.ControllerMustControlMore) > 0 ||
		r.ControllerMustHaveCitysBlessing
}

// unmetDefenderClauseLocked is the first of r's defending-player clauses
// that `defender` does not meet, as the reason a player reads ("Bob
// controls no Island", "Bob isn't poisoned"), or "" when every clause
// holds. `you` is the attacking creature's controller, the "you" of
// "unless you control more creatures than defending player".
//
// Caller must hold g.mu with fresh layers.
func (r AttackTargetRestriction) unmetDefenderClauseLocked(g *Game, you, defender uuid.UUID) string {
	name := g.playerNameLocked(defender)
	if len(r.DefenderMustControl) > 0 && !g.controlsMatchingLocked(defender, r.DefenderMustControl) {
		return name + " controls no " + PermanentQueriesNoun(r.DefenderMustControl)
	}
	p := g.playerByIDLocked(defender)
	if r.DefenderMustBePoisoned && (p == nil || p.Counters[CounterPoison] < 1) {
		return name + " isn't poisoned"
	}
	if r.DefenderMustBeMonarch && g.Monarch != defender {
		return name + " isn't the monarch"
	}
	if n := r.DefenderGraveyardAtLeast; n > 0 && (p == nil || p.Graveyard == nil || p.Graveyard.Size() < n) {
		return name + " has fewer than " + countWord(n) + " cards in their graveyard"
	}
	if qs := r.ControllerMustControlMore; len(qs) > 0 &&
		g.countMatchingLocked(you, qs) <= g.countMatchingLocked(defender, qs) {
		return g.playerNameLocked(you) + " doesn't control more " + permanentQueriesPlural(qs) + " than " + name
	}
	if r.ControllerMustHaveCitysBlessing && !g.CitysBlessingForEffect(you) {
		return g.playerNameLocked(you) + " doesn't have the city's blessing"
	}
	return ""
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
	// Protected is the name of the player a NotPlayer restriction
	// protects, for its sentence.
	Protected string
	// For a refusal by a defending-player clause: TargetName is what was
	// attacked (a player's or a permanent's name), DefenderName the
	// defending player, and Why the clause they do not meet ("Bob
	// controls no Island", "Bob isn't poisoned").
	TargetName   string
	DefenderName string
	Why          string
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
		why := e.Why
		if why == "" {
			why = e.DefenderName + " controls no " + PermanentQueriesNoun(e.Restriction.DefenderMustControl)
		}
		s := e.AttackerName + " can't attack " + e.TargetName + ": " + why
		if e.Restriction.SourceName != "" && e.Restriction.Source != e.Attacker {
			s += " (" + e.Restriction.SourceName + ")"
		}
		return s + "."
	}
	if e.Protected != "" {
		s := e.AttackerName + " can't attack " + e.Protected + " or planeswalkers " + e.Protected + " controls"
		if e.Restriction.NotPlayerOnly {
			s = e.AttackerName + " can't attack " + e.Protected
		}
		if e.Restriction.SourceName != "" && e.Restriction.Source != e.Attacker {
			s += " (" + e.Restriction.SourceName + ")"
		}
		return s + "."
	}
	if e.Restriction.NotAlreadyAttackedThisTurn && !e.Restriction.NotOwner && !e.Restriction.NotOwnersPlaneswalkers {
		return e.AttackerName + " can't attack a player it has already attacked this turn."
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
// for a refusal by a defending-player clause, that defending player and
// the clause they do not meet.
//
// Caller must hold g.mu with fresh layers.
func (r AttackTargetRestriction) refuses(g *Game, attacker *Card, target uuid.UUID) (bool, uuid.UUID, string) {
	if !r.bindsAttackerLocked(g, attacker) {
		return false, uuid.Nil, ""
	}
	if r.ownerRefuses(g, attacker, target) {
		return true, uuid.Nil, ""
	}
	if r.NotAlreadyAttackedThisTurn && g.alreadyAttackedPlayerLocked(attacker, target) {
		return true, uuid.Nil, ""
	}
	if r.notPlayerRefuses(g, target) {
		return true, uuid.Nil, ""
	}
	if r.inactiveLastTurnRefuses(g, target) {
		return true, target, inactiveLastTurnReason(g.playerNameLocked(target))
	}
	if r.asksOfTheDefender() {
		// CR 508.5: the player attacked, a planeswalker's controller,
		// or a battle's protector.
		defender := g.defendingPlayerForAttackLocked(target)
		if defender == uuid.Nil {
			return false, uuid.Nil, ""
		}
		if why := r.unmetDefenderClauseLocked(g, attacker.Controller, defender); why != "" {
			return true, defender, why
		}
	}
	return false, uuid.Nil, ""
}

// bindsAttackerLocked reports whether r applies to `attacker` at all:
// always, unless OnlyAttackersMatching narrows it to attackers with some
// characteristic, read live off the attacker's effective values.
//
// Caller must hold g.mu with fresh layers.
func (r AttackTargetRestriction) bindsAttackerLocked(g *Game, attacker *Card) bool {
	if len(r.OnlyAttackersMatching) == 0 {
		return true
	}
	for _, q := range r.OnlyAttackersMatching {
		if q.matchesLocked(g, attacker) {
			return true
		}
	}
	return false
}

// alreadyAttackedPlayerLocked reports whether `target` is a player this
// object (same epoch) has been declared attacking earlier this turn.
//
// Caller must hold g.mu.
func (g *Game) alreadyAttackedPlayerLocked(attacker *Card, target uuid.UUID) bool {
	if g.classifyAttackTargetLocked(target) != AttackTargetPlayer {
		return false
	}
	for _, p := range g.AttackedPlayersThisTurn(attacker.InstanceID) {
		if p == target {
			return true
		}
	}
	return false
}

// notPlayerRefuses is the NotPlayer clause: the named player, and a
// planeswalker they control (CR 508.1c, Promise of Loyalty).
//
// Caller must hold g.mu with fresh layers.
func (r AttackTargetRestriction) notPlayerRefuses(g *Game, target uuid.UUID) bool {
	if r.NotPlayer == uuid.Nil {
		return false
	}
	switch g.classifyAttackTargetLocked(target) {
	case AttackTargetPlayer:
		return target == r.NotPlayer
	case AttackTargetPlaneswalker:
		if r.NotPlayerOnly {
			return false
		}
		pw := findBattlefieldCard(g, target)
		return pw != nil && pw.Controller == r.NotPlayer
	}
	return false
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
		refused, defender, why := r.refuses(g, attacker, target)
		if !refused {
			continue
		}
		err := &AttackTargetRestrictionError{
			Attacker:     attacker.InstanceID,
			AttackerName: attacker.Effective().Name,
			Restriction:  r,
		}
		if r.NotPlayer != uuid.Nil && r.notPlayerRefuses(g, target) {
			err.Protected = g.playerNameLocked(r.NotPlayer)
		}
		if defender != uuid.Nil {
			err.DefenderName = g.playerNameLocked(defender)
			err.Why = why
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
// a restriction with a defending-player clause, for the card's chip (ADR
// 0107 §2 decision 3): "can't attack Bob: Bob controls no Island".
type DefenderRefusal struct {
	// Player is the opponent who does not meet the restriction. The
	// refusal covers their planeswalkers and the battles they protect
	// too (CR 508.5).
	Player uuid.UUID
	// Restriction is the one that refuses them.
	Restriction AttackTargetRestriction
	// Why is the clause they do not meet, in a player's words: "Bob
	// controls no Island", "Bob isn't the monarch".
	Why string
	// PlayerOnly is set for a refusal that covers the player alone and
	// not their planeswalkers or battles (Arboria's, #1997).
	PlayerOnly bool
}

// DefenderRefusalsForEffect lists, for each restriction on `c` with a
// defending-player clause, every live opponent of its controller who
// does not meet it right now. Nil for nearly every card.
//
// Read-only. Caller must hold g.mu with fresh layers.
func (g *Game) DefenderRefusalsForEffect(c *Card) []DefenderRefusal {
	if c == nil {
		return nil
	}
	var out []DefenderRefusal
	for _, r := range c.Effective().AttackTargetRestrictions {
		if len(r.OnlyAttackersMatching) > 0 && r.NotPlayer != uuid.Nil && r.NotPlayer != c.Controller && r.bindsAttackerLocked(g, c) {
			// #4601: "creatures with flying can't attack you", a fixed
			// player and a filter on the attacker. It is a row only while
			// the filter binds this creature, so one that loses flying
			// drops the chip.
			out = append(out, DefenderRefusal{
				Player: r.NotPlayer, Restriction: r,
				Why:        "it's a " + PermanentQueriesNoun(r.OnlyAttackersMatching),
				PlayerOnly: r.NotPlayerOnly,
			})
		}
		if r.PlayerMustHaveActedLastTurn {
			// #1997: a player clause, not a defending-player one.
			for _, p := range g.Seats {
				if p == nil || p.Eliminated || p.ID == c.Controller || !r.inactiveLastTurnRefuses(g, p.ID) {
					continue
				}
				out = append(out, DefenderRefusal{Player: p.ID, Restriction: r, Why: inactiveLastTurnReason(g.playerNameLocked(p.ID)), PlayerOnly: true})
			}
		}
		if !r.asksOfTheDefender() {
			continue
		}
		for _, p := range g.Seats {
			if p == nil || p.Eliminated || p.ID == c.Controller {
				continue
			}
			if why := r.unmetDefenderClauseLocked(g, c.Controller, p.ID); why != "" {
				out = append(out, DefenderRefusal{Player: p.ID, Restriction: r, Why: why})
			}
		}
	}
	return out
}

// dropSelfAttackTargetRestrictions removes the restrictions `c` imposes
// on itself, for a layer-6 "loses all abilities": each is one of its own
// abilities (CR 613.1f). A restriction another permanent imposes stays.
func dropSelfAttackTargetRestrictions(ch *Characteristic, c *Card) {
	if len(ch.AttackTargetRestrictions) == 0 {
		return
	}
	kept := ch.AttackTargetRestrictions[:0:0]
	for _, r := range ch.AttackTargetRestrictions {
		if r.Source != c.InstanceID {
			kept = append(kept, r)
		}
	}
	ch.AttackTargetRestrictions = kept
}

// canAttackTargetWithLocked reports whether `attacker` may be DECLARED
// attacking `target`: CR 506.2 and 310.9b for its controller
// (canAttackTargetLocked), then its own CR 508.1c target restrictions,
// then defender as its waivers leave it for this target (#1936).
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
	// #2109: the attacking PLAYER's own "can't attack <player>" grant,
	// ahead of the creature's list.
	if err := g.playerCantAttackRefusalLocked(attacker, target); err != nil {
		return err
	}
	// ADR 0206 (#4197): "each player may attack only the nearest
	// opponent in the chosen direction", a rule on every player.
	if err := g.attackDirectionRefusalLocked(attacker, target); err != nil {
		return err
	}
	if err := g.attackTargetRestrictionRefusalLocked(attacker, target); err != nil {
		return err
	}
	// #1936, CR 702.3b: defender, asked of this target, because a
	// waiver may name whom the creature can attack (Weathered
	// Sentinels). Last, as the single verb always refused it after the
	// target checks.
	return g.defenderRefusalLocked(attacker, target)
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
	if len(attacker.Effective().AttackTargetRestrictions) == 0 && !g.playerHasCantAttackGrantLocked(attacker.Controller) &&
		!HasKeyword(attacker, "defender") && !g.attackDirectionInForceLocked(attacker.Controller) {
		return all
	}
	out := all[:0:0]
	for _, t := range all {
		if g.playerCantAttackRefusalLocked(attacker, t.ID) == nil &&
			g.attackDirectionRefusalLocked(attacker, t.ID) == nil &&
			g.attackTargetRestrictionRefusalLocked(attacker, t.ID) == nil &&
			g.defenderRefusalLocked(attacker, t.ID) == nil {
			out = append(out, t)
		}
	}
	return out
}
