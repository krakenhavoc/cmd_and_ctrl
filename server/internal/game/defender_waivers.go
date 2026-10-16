package game

import (
	"github.com/google/uuid"
)

// defender_waivers.go is ADR 0106's 2026-10-16 amendment (#1936): "can
// attack … as though it didn't have defender", asked per target.
//
// Defender is "This creature can't attack" (CR 702.3b). Most cards that
// lift it do so for every target, and the catalog has written those as
// losing the keyword (Stalked Researcher, Ghalta the Immovable). That
// stops working the moment the waiver names WHOM the creature may
// attack. Weathered Sentinels "can attack players who attacked you
// during their last turn as though it didn't have defender": in
// Commander it may attack the opponent who swung at you and nobody else,
// and it still has defender for every other purpose (CR 609.4: an "as
// though" applies only to the stated effect). So the waiver is data on
// the characteristic, asked of each target where the defender check is
// made, the way AttackTargetRestrictions are:
//
//   - AttackerEligible, the attacker-side list both the bulk declaration
//     verb and the enumerator run, lets a creature with defender through
//     only when it has at least one waiver;
//   - canAttackTargetWithLocked, the per-target predicate both
//     declaration verbs run, refuses a target no waiver covers;
//   - AttackTargetsForAttackerForEffect, the enumerator's per-attacker
//     list, drops those targets, so the ADR 0105 digest's attack_targets,
//     the client's per-attacker gate and the bot's moves follow (#544);
//   - attackRequirementCandidatesLocked, the CR 508.1d search, counts a
//     goaded Sentinels as able to attack only the players it may attack.
//
// CR 508.4c's "put onto the battlefield attacking" and CR 508.7b's
// reselection never asked about defender, and still do not.
//
// The same field carries the unconditional form, written by the
// ScopedEffect mod kind ModWaiveDefender: Hightide Hermit's "This
// creature can attack this turn as though it didn't have defender",
// which, unlike losing the keyword, leaves the creature a creature with
// defender for anything that counts one.

// DefenderWaiver is one "can attack as though it didn't have defender"
// on one creature. Pure data, so the Characteristic that carries it stays
// copyable and the snapshot's lastKnownBattlefield round-trips it.
type DefenderWaiver struct {
	// Source is the permanent whose ability or effect it is, and
	// SourceName that permanent's name for the refusal sentence.
	Source     uuid.UUID
	SourceName string
	// OnlyPlayersWhoAttackedYou limits it to attacking a player who
	// attacked the creature's controller during that player's last turn
	// (Weathered Sentinels; ADR 0108 §6's record). Players only: a
	// planeswalker or battle is not "a player", so the creature may not
	// attack one under this waiver. Unset, the waiver covers every
	// target.
	OnlyPlayersWhoAttackedYou bool `json:",omitempty"`
}

// waivesLocked reports whether w lets `attacker` attack `target` as
// though it didn't have defender. "You" is the attacker's controller,
// read live: a stolen Sentinels answers to its new controller's history.
//
// Caller must hold g.mu with fresh layers.
func (w DefenderWaiver) waivesLocked(g *Game, attacker *Card, target uuid.UUID) bool {
	if !w.OnlyPlayersWhoAttackedYou {
		return true
	}
	if g.classifyAttackTargetLocked(target) != AttackTargetPlayer {
		return false
	}
	return g.AttackedYouDuringTheirLastTurn(target, attacker.Controller)
}

// hasDefenderWaiver reports whether c can attack anything at all as
// though it didn't have defender. The attacker-side half: which targets
// is defenderRefusalLocked's question.
func hasDefenderWaiver(c *Card) bool {
	return c != nil && len(c.Effective().DefenderWaivers) > 0
}

// DefenderWaiverError is the refusal for a creature with defender whose
// waivers do not reach the target: "Weathered Sentinels has defender: it
// can attack only players who attacked you during their last turn." It
// wraps ErrDefender, so every caller that tests for that keeps working.
type DefenderWaiverError struct {
	AttackerName string
	// Waiver is the first waiver the creature has, for the sentence.
	Waiver DefenderWaiver
}

func (e *DefenderWaiverError) Error() string { return "game: " + e.Sentence() }

// Unwrap makes errors.Is(err, ErrDefender) hold.
func (e *DefenderWaiverError) Unwrap() error { return ErrDefender }

// Sentence is the refusal in a player's words.
func (e *DefenderWaiverError) Sentence() string {
	if e.Waiver.OnlyPlayersWhoAttackedYou {
		return e.AttackerName + " has defender: it can attack only players who attacked you during their last turn."
	}
	return e.AttackerName + " has defender and can't attack."
}

// defenderRefusalLocked is CR 702.3b asked of one target: nil when
// `attacker` lacks defender or one of its waivers covers `target`,
// ErrDefender for a creature with defender and no waiver, and a
// *DefenderWaiverError when it has waivers and none covers this target.
//
// Caller must hold g.mu with fresh layers.
func (g *Game) defenderRefusalLocked(attacker *Card, target uuid.UUID) error {
	if attacker == nil || !HasKeyword(attacker, "defender") {
		return nil
	}
	waivers := attacker.Effective().DefenderWaivers
	if len(waivers) == 0 {
		return ErrDefender
	}
	for _, w := range waivers {
		if w.waivesLocked(g, attacker, target) {
			return nil
		}
	}
	return &DefenderWaiverError{AttackerName: attacker.Effective().Name, Waiver: waivers[0]}
}
