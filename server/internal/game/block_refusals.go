package game

import (
	"errors"

	"github.com/google/uuid"
)

// block_refusals.go — #3898: why a creature can't block an attacker,
// asked ahead of the click.
//
// The client declares blockers on the board: click your creature, then
// the attacker it blocks. While a creature is selected, the attackers it
// can't block are drawn unclickable with the reason on hover, and an
// attacker that takes more than one blocker (menace) collects a second
// one before anything is sent. Both answers are the verb's own: each
// pairing is put through the same validator DeclareBlockers runs
// (checkBlockDeclarationLocked), on top of the declaration as it stands,
// so the board never greys an attacker the engine would accept or offers
// one it would refuse (ADR 0045 §3, ADR 0105 §1).

// LoneBlockRefusal is one attacker a creature could not block on its
// own right now, and why.
type LoneBlockRefusal struct {
	Attacker uuid.UUID
	Reason   BlockReason
	// Sentence is the refusal's player-facing sentence, addressed to the
	// blocker's controller (BlockRefusedError.Sentence).
	Sentence string
}

// LoneBlockRefusalsForEffect lists, for every creature a defending
// player who is still declaring could put into the declaration — one
// that may block now (BlockerEligible) or one already blocking, which a
// click may re-point — each attacker that player defends against that
// the creature could not be declared blocking ALONE, with the
// validator's reason. A lone block on a menace attacker is refused
// with too_few_blockers: the client reads that, with the attacker's
// MinBlockersForEffect, as "pick another blocker too".
//
// Pairings the declaration already holds are skipped (clicking one
// takes the block back), and so is anything face down: a refusal's
// sentence names both creatures. Empty outside declare_blockers.
//
// Caller must hold g.mu (read or write) with fresh layers. Reads only.
func (g *Game) LoneBlockRefusalsForEffect() map[uuid.UUID][]LoneBlockRefusal {
	if g.Battlefield == nil || g.Turn.Step != StepDeclareBlockers {
		return nil
	}
	var out map[uuid.UUID][]LoneBlockRefusal
	base := g.currentBlockAssignmentLocked()
	cache := blockReachCache{}
	for _, s := range g.Seats {
		if s == nil || s.Eliminated || g.BlockDeclarationStatusLocked(s.ID) != BlockDeclarationPending {
			continue
		}
		var attackers []*Card
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			if c.AttackingTarget == uuid.Nil || c.FaceDown {
				continue
			}
			if g.defendingPlayerForAttackerLocked(c) == s.ID {
				attackers = append(attackers, c)
			}
		}
		if len(attackers) == 0 {
			continue
		}
		for i := range g.Battlefield.Cards {
			b := &g.Battlefield.Cards[i]
			if b.FaceDown || b.Controller != s.ID || !b.IsCreature() {
				continue
			}
			if !BlockerEligible(b, s.ID) && len(b.BlockedAttackers()) == 0 {
				continue
			}
			for _, atk := range attackers {
				if base.has(b.InstanceID, atk.InstanceID) {
					continue
				}
				decls := []BlockDeclaration{{Blocker: b.InstanceID, Attacker: atk.InstanceID}}
				err := g.checkBlockDeclarationCachedLocked(base, decls, cache)
				var refused *BlockRefusedError
				if err == nil || !errors.As(err, &refused) {
					continue
				}
				if out == nil {
					out = map[uuid.UUID][]LoneBlockRefusal{}
				}
				out[b.InstanceID] = append(out[b.InstanceID], LoneBlockRefusal{
					Attacker: atk.InstanceID,
					Reason:   refused.Reason,
					Sentence: refused.Sentence(s.ID),
				})
			}
		}
	}
	return out
}

// MinBlockersForEffect is the fewest creatures that may block
// `attacker` (CR 509.1b): 2 for menace, more for a rule that says so,
// while the player defending against it is still declaring blockers;
// 0 when there is no minimum, or no declaration it could matter to.
// The client reads it to know how many blockers a group needs before it
// is sent, and which blocks go with a blocker taken back out of one.
//
// Caller must hold g.mu (read or write) with fresh layers.
func (g *Game) MinBlockersForEffect(attacker *Card) int {
	if attacker == nil || attacker.AttackingTarget == uuid.Nil || g.Turn.Step != StepDeclareBlockers {
		return 0
	}
	if g.BlockDeclarationStatusLocked(g.defendingPlayerForAttackerLocked(attacker)) != BlockDeclarationPending {
		return 0
	}
	min, _ := g.blockerBoundsLocked(attacker)
	return min
}
