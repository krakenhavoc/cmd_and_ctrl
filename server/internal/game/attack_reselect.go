package game

import (
	"errors"

	"github.com/google/uuid"
)

// attack_reselect.go — CR 508.7, "reselect which player, planeswalker,
// or battle a creature is attacking" (#1329, ADR 0045 amendment
// 2026-09-23).
//
// The rules, verbatim (August 7, 2026):
//
//	508.7   Some cards allow a player to reselect which player,
//	        planeswalker, or battle a creature is attacking.
//	508.7a  The attacking creature isn't removed from combat and it
//	        isn't considered to have attacked a second time. That
//	        creature is attacking the reselected player or permanent,
//	        but it's still considered to have attacked the player or
//	        permanent chosen as it was declared as an attacker.
//	508.7b  While reselecting which player, planeswalker, or battle a
//	        creature is attacking, that creature isn't affected by
//	        requirements or restrictions that apply to the declaration
//	        of attackers.
//	508.7c  The reselected player, planeswalker, or battle must be an
//	        opponent of the attacking creature's controller, a
//	        planeswalker controlled by an opponent of the attacking
//	        creature's controller, or a battle protected by an opponent
//	        of the attacking creature's controller.
//	508.7d  In a multiplayer game not using the attack multiple players
//	        option …
//
// # What each clause is, here
//
//   - 508.7a is what this verb does NOT do. It writes
//     Card.AttackingTarget and nothing else: no EventAttack (the
//     creature did not attack again, so no "whenever ~ attacks"
//     trigger and no Adeline batch), no removal from combat, no
//     change to announcedAttacks (the creature WAS announced, and
//     stays so), and no change to the blocked record — a creature
//     that was blocked is still blocked (CR 509.1h) by the same
//     blockers. The triggers the declaration already produced were
//     harvested off the defender it was declared against, which is
//     the second sentence of 508.7a, and is true for free: the
//     declaration's lock-in (Decision 22) already happened.
//   - 508.7b is why this does not go through DeclareAttackerWith: no
//     summoning-sickness or defender check, no tap, no CR 508.1a
//     restriction, and no Propaganda tax (attack_tax.go charges the
//     DECLARATION, which this is not).
//   - 508.7c is canAttackTargetLocked, read against the attacking
//     creature's controller — the same function the declaration uses,
//     because 508.7c and CR 506.2 name the same set.
//   - 508.7d does not apply. CR 903.2: a Commander game's default
//     multiplayer setup is Free-for-All WITH the attack multiple
//     players option, which is the only setup this engine plays.
//
// The rest of combat reads AttackingTarget live, so it follows the
// new target with no other change: defendingPlayerForAttackLocked
// makes the new defender's creatures the legal blockers in the
// declare-blockers step, and the combat damage steps deal an
// unblocked creature's damage to the new target (CR 510.1b).

// ErrNotAttacking is returned by ReselectAttackTargetForEffect when
// the named permanent is not an attacking creature: not on the
// battlefield, not a creature, or not in combat as an attacker.
var ErrNotAttacking = errors.New("game: that creature is not attacking")

// ReselectAttackTargetForEffect makes `attacker` attack `target`
// instead of what it is attacking now (CR 508.7). See the file header
// for what it deliberately does not do.
//
// "Attacking" means in combat as an attacker: AttackingTarget is set
// AND the creature has had its announcement — declared and locked in
// (Decision 22), or put onto the battlefield attacking (CR 506.3c,
// stampEntryAttackerLocked). A creature that is only STAGED — clicked
// in the declare-attackers step, declaration not complete — is
// refused: its declaration is still the active player's to change,
// and a reselection landing before the lock-in would make the lock-in
// announce the reselected defender, which CR 508.7a forbids. No
// effect can resolve in that window (the lock-in runs before anyone
// receives priority), so the refusal only ever reaches a caller that
// is not a resolving effect.
//
// An attacker whose planeswalker or battle has left combat still
// counts as attacking (CR 506.4c — "it continues to be an attacking
// creature, although it is not attacking any player, planeswalker, or
// battle"), and may be reselected onto something real.
//
// Reselecting the target it already has is a successful no-op.
//
// Caller must hold g.mu.
func (g *Game) ReselectAttackTargetForEffect(attacker, target uuid.UUID) error {
	c := findBattlefieldCard(g, attacker)
	if c == nil || !c.IsCreature() || c.AttackingTarget == uuid.Nil || !g.announcedAttacks[attacker] {
		return ErrNotAttacking
	}
	if err := g.canAttackTargetLocked(c.Controller, target); err != nil {
		return err
	}
	if c.AttackingTarget == target {
		return nil
	}
	c.AttackingTarget = target
	// "Is this creature attacking YOU" just changed for two players
	// with no EventAttack to say so — the same gap #1218 closed for a
	// control change and for combat ending.
	g.invalidateLayersForAttackChangeLocked()
	return nil
}

// ReselectAttackTargetsForEffect lists what `attacker` may be
// reselected onto (CR 508.7c): AttackTargetsForEffect for its
// controller, less the target it already has. Empty when it is not an
// attacking creature, or when there is nothing else it could attack.
//
// Caller must hold g.mu.
func (g *Game) ReselectAttackTargetsForEffect(attacker uuid.UUID) []AttackTargetRef {
	c := findBattlefieldCard(g, attacker)
	if c == nil || !c.IsCreature() || c.AttackingTarget == uuid.Nil || !g.announcedAttacks[attacker] {
		return nil
	}
	var out []AttackTargetRef
	for _, ref := range g.AttackTargetsForEffect(c.Controller) {
		if ref.ID != c.AttackingTarget {
			out = append(out, ref)
		}
	}
	return out
}

// ReselectAttackPrompt is the queue-side description of "you may
// reselect which player or permanent that creature is attacking".
type ReselectAttackPrompt struct {
	// Chooser answers. Required: the controller of the effect doing
	// the reselecting, who is frequently NOT the attacker's controller
	// (a defending player flashing in Misleading Signpost).
	Chooser uuid.UUID

	// Attacker is the attacking creature being redirected.
	Attacker uuid.UUID

	// Source is the card asking.
	Source uuid.UUID

	// Question is the prompt's header, written the way the card is.
	Question string

	// Then runs once the question is settled — answered, declined, or
	// dropped — and is how a card that asks once per attacking
	// creature (Windshaper Planetar) chains the next one. It also runs,
	// inline, when there was nothing to ask. Runs with g.mu held.
	Then func(g *Game) error
}

// QueueReselectAttackForEffect asks p.Chooser whether to reselect
// p.Attacker's attack (CR 508.7) and, if so, onto what. Returns the
// prompt's ID, or uuid.Nil when nothing was queued — the attacker is
// not attacking, or there is nothing else it could attack — in which
// case p.Then has already run.
//
// ONE option_pick, and the "you may" is its FIRST option: "Keep
// attacking <current>". That is the branch the enumerator marks
// always-legal, which is right twice over — declining is always
// possible, and a bot that has no opinion leaves somebody else's combat
// alone. The rest are the legal targets, a seat as a Player option and
// a planeswalker or battle as a card option, so the client renders the
// permanent rather than quoting its name.
//
// The answer is read off the OPTION the chooser picked, not an index
// into a list captured here (optionPickFrame.thenSubject): a seat that
// leaves the game is pruned off an open prompt (#994), which renumbers
// the options after it. The picked target is then re-validated by
// ReselectAttackTargetForEffect, so an answer the board has since made
// illegal changes nothing rather than erroring.
//
// Caller must hold g.mu.
func (g *Game) QueueReselectAttackForEffect(p ReselectAttackPrompt) uuid.UUID {
	then := p.Then
	finish := func(g *Game) error {
		if then == nil {
			return nil
		}
		return then(g)
	}
	refs := g.ReselectAttackTargetsForEffect(p.Attacker)
	c := findBattlefieldCard(g, p.Attacker)
	if len(refs) == 0 || c == nil {
		if err := finish(g); err != nil {
			g.emitChoiceEffectErrorLocked(p.Chooser, p.Source, err)
		}
		return uuid.Nil
	}
	options := make([]ChoiceOption, 0, len(refs)+1)
	options = append(options, ChoiceOption{
		Label: "Keep attacking " + g.attackTargetLabelLocked(c.AttackingTarget),
	})
	for _, ref := range refs {
		opt := ChoiceOption{Label: g.attackTargetLabelLocked(ref.ID)}
		if ref.Kind == AttackTargetPlayer {
			opt.Player = ref.ID
		} else {
			opt.Cards = []uuid.UUID{ref.ID}
		}
		options = append(options, opt)
	}
	attacker := p.Attacker
	queued := g.QueueChoiceForEffect(PendingChoice{
		Kind:        PendingChoiceOptionPick,
		Chooser:     p.Chooser,
		FromPlayer:  p.Chooser,
		Count:       1,
		Source:      p.Source,
		Reason:      p.Question,
		PickOptions: options,
		optionPickResume: &optionPickFrame{
			thenSubject: func(g *Game, target uuid.UUID) error {
				if target != uuid.Nil {
					// An illegal answer — the target left, or its
					// controller did — reselects nothing. The
					// question was "you may", and "no" is always a
					// legal outcome of it.
					if err := g.ReselectAttackTargetForEffect(attacker, target); err != nil &&
						!errors.Is(err, ErrNotAttacking) && !errors.Is(err, ErrIllegalAttackTarget) {
						return err
					}
				}
				return finish(g)
			},
		},
	})
	if queued == uuid.Nil {
		if err := finish(g); err != nil {
			g.emitChoiceEffectErrorLocked(p.Chooser, p.Source, err)
		}
	}
	return queued
}

// attackTargetLabelLocked names an attack target on an option button:
// a seat's name, or a permanent's name with the seat that defends it —
// "Jace, the Mind Sculptor (Bob)", since two players can control two
// copies of one planeswalker. A target that no longer resolves (a
// planeswalker removed from combat, CR 506.4c) reads as "nothing".
// Caller must hold g.mu.
func (g *Game) attackTargetLabelLocked(id uuid.UUID) string {
	if p := g.playerByIDLocked(id); p != nil {
		return g.seatLabelLocked(id)
	}
	if c := findBattlefieldCard(g, id); c != nil {
		if defender := g.defendingPlayerForAttackLocked(id); defender != uuid.Nil {
			return c.Name + " (" + g.seatLabelLocked(defender) + ")"
		}
		return c.Name
	}
	return "nothing"
}
