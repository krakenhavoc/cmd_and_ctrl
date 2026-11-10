package game

import "github.com/google/uuid"

// block_not_alone.go is "This creature can't block alone" (#5396: the
// 4/4 Beast token Toby, Beastie Befriender makes, which "can't attack or
// block alone"): BlockRule.NotAlone. The attack half is
// AttackLimitNotAlone (attack_not_alone.go).
//
// CR 506.5: "A creature blocks alone if it's the only creature declared
// as a blocker during the declare blockers step." CR 509.1b makes the
// restriction a property of the declaration: "If any restrictions are
// being disobeyed, the declaration of blockers is illegal." And in a
// multiplayer game CR 802.4b says "When determining whether a defending
// player's blocks are legal, ignore any creatures attacking other
// players and any blocking creatures controlled by other players." So
// "alone" is counted among the blockers ONE defending player declares:
// another player's blocker is no company.
//
// Nothing here forbids the creature; it forbids a declaration whose one
// blocker it is. Two creatures that each can't block alone may block
// together (the block-side reading of CR 508.1c's example), on the same
// attacker or on two.
//
// Judged in two places, like the attack half:
//
//   - the verb (blockAloneRefusalLocked, the validator's last
//     restriction check): a declaration that would leave the creature
//     as its controller's only blocker is refused with
//     cant_block_alone, so the engine never holds a lone block it would
//     refuse (ADR 0045 Decision 13). Another blocker declared first, or
//     both in one declare_blockers, is legal. The option generator, the
//     #328 signal and the view's block_refusals all run the same
//     validator, so none of them offers what the verb refuses.
//   - the completion (takeOutLoneNotAloneBlockerLocked): a declaration
//     can still be left with the creature alone by something the
//     validator does not judge — the sandbox's remove_from_combat
//     taking its companion back, or the companion leaving the
//     battlefield before the defender finishes. The declaration the
//     defender completes would be illegal, so the creature is taken out
//     of it before anything is announced, and the defender has declared
//     no blocks.

// cantBlockAloneLocked reports whether `c` prints (or has been given
// through its catalog entry) "this creature can't block alone". Keyed
// by CatalogAbilityKey, so a creature that has lost all its abilities
// may block alone (CR 613.1f).
//
// Caller must hold g.mu (read or write). Reads only.
func cantBlockAloneLocked(c *Card) bool {
	if c == nil || CatalogBlockRules == nil {
		return false
	}
	key := catalogAbilityKeyOf(c)
	if key == "" {
		return false
	}
	for _, r := range CatalogBlockRules(key) {
		if r.NotAlone {
			return true
		}
	}
	return false
}

// CantBlockAlone reports whether `c` can't block alone (BlockRule.NotAlone,
// #5396): the view's cant_block_alone stamp, which the bot's block search
// reads to keep such a creature out of a block set it would be alone in.
// Reads the catalog only; safe under any lock the caller holds.
func CantBlockAlone(c *Card) bool { return cantBlockAloneLocked(c) }

// loneBlockerOfLocked is the one creature `seat` blocks with in
// `assign`, or nil when it blocks with none or with more than one.
//
// Caller must hold g.mu (read or write). Reads only.
func (g *Game) loneBlockerOfLocked(seat uuid.UUID, assign blockAssignment) *Card {
	var lone *Card
	for id, atks := range assign {
		if len(atks) == 0 {
			continue
		}
		c := findBattlefieldCard(g, id)
		if c == nil || c.Controller != seat {
			continue
		}
		if lone != nil {
			return nil
		}
		lone = c
	}
	return lone
}

// blockAloneRefusalLocked is the validator's "can't block alone" check:
// for each defending player whose blockers `decls` names, the refusal
// when `after` leaves them blocking with exactly one creature and that
// creature can't block alone, or nil.
//
// Only the players the declaration touches are judged, so a lone block
// the validator never saw (its companion taken back by the sandbox) does
// not make every other player's block illegal; the completion takes
// that one out.
//
// Caller must hold g.mu with fresh layers. Reads only.
func (g *Game) blockAloneRefusalLocked(after blockAssignment, decls []BlockDeclaration) *BlockRefusedError {
	if CatalogBlockRules == nil {
		return nil
	}
	var seen []uuid.UUID
	for _, d := range decls {
		b := findBattlefieldCard(g, d.Blocker)
		if b == nil {
			continue
		}
		judged := false
		for _, s := range seen {
			if s == b.Controller {
				judged = true
				break
			}
		}
		if judged {
			continue
		}
		seen = append(seen, b.Controller)
		lone := g.loneBlockerOfLocked(b.Controller, after)
		if lone == nil || !cantBlockAloneLocked(lone) {
			continue
		}
		var attacker *Card
		if atks := after[lone.InstanceID]; len(atks) > 0 {
			attacker = findBattlefieldCard(g, atks[0])
		}
		return g.blockRefusedErrorLocked(attacker, lone,
			BlockRefusal{Reason: BlockReasonCantBlockAlone, Source: lone.InstanceID})
	}
	return nil
}

// takeOutLoneNotAloneBlockerLocked is the completion's half: when
// `seat`'s staged declaration is one creature that can't block alone,
// that creature is taken out of combat before the declaration is
// announced (CR 509.1b: the declaration it was in is illegal). Nothing
// was announced for it yet, so nothing is taken back that anyone saw.
//
// Caller must hold g.mu in write mode.
func (g *Game) takeOutLoneNotAloneBlockerLocked(seat uuid.UUID) {
	if CatalogBlockRules == nil || g.Battlefield == nil {
		return
	}
	lone := g.loneBlockerOfLocked(seat, g.currentBlockAssignmentLocked())
	if lone == nil || !cantBlockAloneLocked(lone) {
		return
	}
	g.removeFromCombatLocked(lone)
}

// notAloneBlockerGroupsLocked is the option generator's addition for a
// creature that can't block alone, offered while `seat` has staged no
// block: the creature and one companion, both on `atk`, as one
// declaration. Without it, the creature's only way into a declaration is
// a second move after a companion's, which a policy that scores one move
// at a time never sees. Companions on another attacker stay reachable as
// that order of singles. Up to `limit` groups, in battlefield order;
// each is put through the validator.
//
// Caller must hold g.mu with fresh layers. Reads only.
func (g *Game) notAloneBlockerGroupsLocked(seat uuid.UUID, atk *Card, pool []*Card, base blockAssignment, cache blockReachCache, limit int) [][]BlockDeclaration {
	if CatalogBlockRules == nil || limit <= 0 || len(pool) < 2 {
		return nil
	}
	for id, atks := range base {
		if len(atks) == 0 {
			continue
		}
		if c := findBattlefieldCard(g, id); c != nil && c.Controller == seat {
			return nil
		}
	}
	var out [][]BlockDeclaration
	for i, b := range pool {
		if !cantBlockAloneLocked(b) {
			continue
		}
		for j, other := range pool {
			// Two creatures that each can't block alone pair once,
			// from the first of them.
			if j == i || (j < i && cantBlockAloneLocked(other)) {
				continue
			}
			decls := []BlockDeclaration{
				{Blocker: b.InstanceID, Attacker: atk.InstanceID},
				{Blocker: other.InstanceID, Attacker: atk.InstanceID},
			}
			if g.checkBlockDeclarationCachedLocked(base, decls, cache) != nil {
				continue
			}
			out = append(out, decls)
			if len(out) >= limit {
				return out
			}
		}
	}
	return out
}
