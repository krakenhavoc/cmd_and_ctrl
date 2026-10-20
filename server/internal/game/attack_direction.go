package game

import "github.com/google/uuid"

// attack_direction.go — ADR 0206 (#4197): "each player may attack only
// the nearest opponent in the chosen direction and planeswalkers
// controlled by that opponent" (Pramikon, Sky Rampart; Teyo, Geometric
// Tactician's −2). A CR 508.1c restriction on every PLAYER's
// declaration of attackers, keyed to a direction at the table.
//
// # Left and right
//
// The table's seats are its turn order (Seats, advanced one index at a
// time by the turn machine), and CR 101.4 puts the next player in turn
// order to a player's left (CR 103.1: turn order proceeds clockwise).
// So "left" is the next seat in turn order and "right" the previous
// one. The nearest opponent in a direction is the first player still in
// the game walking that way: a seat whose player has left the game is
// skipped (CR 800.4a: they are not a player any more, let alone an
// opponent). The engine has no teams, so every other player is an
// opponent (CR 806.1: players compete as individuals). With two players
// left, both directions name the same opponent.
//
// # What may be attacked
//
// The nearest opponent and planeswalkers they control. Nothing else:
// not another opponent, not their planeswalkers, and no battle at all,
// because "may attack only" lists what is allowed and battles are not
// on the list. The opponent is worked out LIVE for the attacking
// creature's controller each time the question is asked, so a player
// leaving the game moves everybody's nearest opponent at once.
//
// # Two sources, as every player-level state has
//
//   - DERIVED: a permanent whose static says so, CardDef.
//     AttackOnlyNearestOpponentInChosenDirection (Pramikon), with the
//     direction read off its CR 614.12 "as this enters, choose left or
//     right" answer (Card.ChosenOption, "Left" or "Right"). It is the
//     permanent's ability, read through its ability key, so one that has
//     lost its abilities restricts nothing; and a permanent with no
//     answer (a copy that entered without choosing) restricts nothing,
//     which is the Pramikon ruling.
//   - GRANTED: PlayerStatic.AttackDirection with a CR 611.2 Duration,
//     written on every player by a resolved ability (Teyo's −2, "until
//     your next turn"). Stored on each player rather than once on the
//     controller, so the controller leaving the game does not take it
//     with them (CR 800.4m keeps the duration running until that turn
//     would have begun).
//
// Every rule in force applies at once. Two that disagree about the
// direction leave a player of a three-or-more-player game nothing to
// attack, which is the rulings' answer for both cards ("players can't
// attack at all"). Teyo's "the last chosen direction" is the direction
// chosen as that ability resolved, not a newer choice replacing an
// older one: the Teyo ruling says a Teyo effect and a similar one that
// disagree leave nobody able to attack.
//
// # Where it is read
//
// attackDirectionRefusalLocked sits in canAttackTargetWithLocked beside
// the attacking player's other grants, so both declaration verbs, the
// CR 508.1d requirement search, and the enumerator's per-attacker list
// (AttackTargetsForAttackerForEffect, which the ADR 0105 digest, the
// client's attack gate and the bot read) all follow. CR 508.4c and
// 508.7b keep the controller-only check: a creature put onto the
// battlefield attacking is not affected (the Teyo ruling says so too).

// AttackDirection is a direction at the table: left (the next player
// in turn order) or right (the previous one). The zero value says
// nothing.
type AttackDirection string

const (
	// AttackDirectionLeft is toward the next player in turn order.
	AttackDirectionLeft AttackDirection = "left"
	// AttackDirectionRight is toward the previous player in turn order.
	AttackDirectionRight AttackDirection = "right"
)

// The anchor words a "choose left or right" prompt offers, in printed
// order. Card.ChosenOption holds one of them on a permanent that chose.
const (
	AttackDirectionLeftWord  = "Left"
	AttackDirectionRightWord = "Right"
)

// AttackDirectionOfWord is the direction an anchor word names, or ""
// for any other word (an unanswered choice included).
func AttackDirectionOfWord(word string) AttackDirection {
	switch word {
	case AttackDirectionLeftWord:
		return AttackDirectionLeft
	case AttackDirectionRightWord:
		return AttackDirectionRight
	}
	return ""
}

// valid reports whether d names a direction.
func (d AttackDirection) valid() bool {
	return d == AttackDirectionLeft || d == AttackDirectionRight
}

// attackDirectionRule is one rule in force: a direction and what said so.
type attackDirectionRule struct {
	Direction  AttackDirection
	SourceName string
}

// GrantAttackDirectionForEffect makes "each player may attack only the
// nearest opponent in direction `dir` and planeswalkers controlled by
// that opponent" true for duration d: one PlayerStatic on every player
// still in the game. Build d with the constructors in duration.go (Teyo
// passes UntilYourNextTurnDuration of its controller). A no-op for a
// direction that is not left or right.
//
// The *ForEffect surface: caller must hold g.mu (write), which a
// resolution frame already does.
func (g *Game) GrantAttackDirectionForEffect(dir AttackDirection, label string, source uuid.UUID, d Duration) {
	if !dir.valid() {
		return
	}
	for _, p := range g.Seats {
		if p == nil || p.Eliminated {
			continue
		}
		p.Statics = append(p.Statics, PlayerStatic{
			AttackDirection: dir,
			Source:          source,
			Label:           label,
			Duration:        d,
		})
	}
}

// attackDirectionRulesLocked lists every direction rule `player`'s
// attacks are under right now: each permanent's static, then each live
// grant on the player. The duration is tested here as well as in the
// sweep, for the reason playerCantAttackRefusalLocked gives.
//
// Caller must hold g.mu (read or write).
func (g *Game) attackDirectionRulesLocked(player uuid.UUID) []attackDirectionRule {
	var out []attackDirectionRule
	if g.Battlefield != nil {
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			dir := AttackDirectionOfWord(c.ChosenOption)
			if !dir.valid() {
				continue
			}
			// The ability key: a Pramikon that has lost its abilities
			// restricts nothing (CR 613.1f).
			def := catalogDef(catalogAbilityKeyOf(c))
			if def == nil || !def.AttackOnlyNearestOpponentInChosenDirection {
				continue
			}
			out = append(out, attackDirectionRule{Direction: dir, SourceName: EffectiveOf(c).Name})
		}
	}
	if p := g.playerByIDLocked(player); p != nil {
		for _, s := range p.Statics {
			if !s.AttackDirection.valid() || g.durationExpiredLocked(s.Duration, false) {
				continue
			}
			out = append(out, attackDirectionRule{Direction: s.AttackDirection, SourceName: s.Label})
		}
	}
	return out
}

// NearestOpponentForEffect is `player`'s nearest opponent in direction
// `dir`: the first player still in the game walking from their seat
// that way round the table, or uuid.Nil when there is none.
//
// Read-only. Caller must hold g.mu.
func (g *Game) NearestOpponentForEffect(player uuid.UUID, dir AttackDirection) uuid.UUID {
	n := len(g.Seats)
	at := -1
	for i, p := range g.Seats {
		if p != nil && p.ID == player {
			at = i
			break
		}
	}
	if at < 0 || !dir.valid() {
		return uuid.Nil
	}
	step := 1
	if dir == AttackDirectionRight {
		step = n - 1
	}
	for k := 1; k < n; k++ {
		p := g.Seats[(at+k*step)%n]
		if p == nil || p.Eliminated || p.ID == player {
			continue
		}
		return p.ID
	}
	return uuid.Nil
}

// attackDirectionAllowsLocked reports whether `target` is the player
// `nearest` or a planeswalker they control. Battles are never allowed:
// "may attack only" does not list them.
//
// Caller must hold g.mu.
func (g *Game) attackDirectionAllowsLocked(target, nearest uuid.UUID) bool {
	if nearest == uuid.Nil {
		return false
	}
	switch g.classifyAttackTargetLocked(target) {
	case AttackTargetPlayer:
		return target == nearest
	case AttackTargetPlaneswalker:
		pw := findBattlefieldCard(g, target)
		return pw != nil && pw.Controller == nearest
	}
	return false
}

// AttackDirectionError is the refusal a declaration verb returns when a
// creature is pointed at something other than its controller's nearest
// opponent in a direction in force (or that opponent's planeswalkers).
// It wraps ErrIllegalAttackTarget, so every caller that tests for that
// keeps working, and its message is the toast's sentence.
type AttackDirectionError struct {
	// Attacker is the creature, AttackerName its name.
	Attacker     uuid.UUID
	AttackerName string
	// Direction is the rule that refused it, and SourceName what said
	// so ("Pramikon, Sky Rampart").
	Direction  AttackDirection
	SourceName string
	// Nearest is the opponent it may attack, NearestName their name;
	// uuid.Nil when there is no opponent that way.
	Nearest     uuid.UUID
	NearestName string
}

func (e *AttackDirectionError) Error() string { return "game: " + e.Sentence() }

// Unwrap makes errors.Is(err, ErrIllegalAttackTarget) hold.
func (e *AttackDirectionError) Unwrap() error { return ErrIllegalAttackTarget }

// Sentence is the refusal in a player's words: "Grizzly Bears may
// attack only Bob, the nearest opponent to the left, and planeswalkers
// Bob controls (Pramikon, Sky Rampart)."
func (e *AttackDirectionError) Sentence() string {
	var s string
	if e.Nearest == uuid.Nil {
		s = e.AttackerName + " can't attack: there is no opponent to the " + string(e.Direction)
	} else {
		s = e.AttackerName + " may attack only " + e.NearestName + ", the nearest opponent to the " +
			string(e.Direction) + ", and planeswalkers " + e.NearestName + " controls"
	}
	if e.SourceName != "" {
		s += " (" + e.SourceName + ")"
	}
	return s + "."
}

// attackDirectionRefusalLocked is THE reader: nil when every direction
// rule in force lets `attacker`'s controller attack `target`, else the
// refusal naming the first that does not.
//
// Caller must hold g.mu with fresh layers.
func (g *Game) attackDirectionRefusalLocked(attacker *Card, target uuid.UUID) error {
	if attacker == nil {
		return nil
	}
	for _, r := range g.attackDirectionRulesLocked(attacker.Controller) {
		nearest := g.NearestOpponentForEffect(attacker.Controller, r.Direction)
		if g.attackDirectionAllowsLocked(target, nearest) {
			continue
		}
		err := &AttackDirectionError{
			Attacker:     attacker.InstanceID,
			AttackerName: attacker.Effective().Name,
			Direction:    r.Direction,
			SourceName:   r.SourceName,
			Nearest:      nearest,
		}
		if nearest != uuid.Nil {
			err.NearestName = g.playerNameLocked(nearest)
		}
		return err
	}
	return nil
}

// attackDirectionInForceLocked is the cheap early-out for the
// enumerator's per-attacker list: is any direction rule in force for
// `player`'s attacks?
//
// Caller must hold g.mu.
func (g *Game) attackDirectionInForceLocked(player uuid.UUID) bool {
	return len(g.attackDirectionRulesLocked(player)) > 0
}
