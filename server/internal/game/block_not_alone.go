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

// blockCompanionRuleLocked is the companion rule `c` prints (or has
// been given through its catalog entry): "this creature can't block
// alone" needs one other blocker, Orcish Conscripts two, Okk one with
// greater power (#6197). ok is false when it prints none. Keyed by
// CatalogAbilityKey, so a creature that has lost all its abilities may
// block alone (CR 613.1f).
//
// Caller must hold g.mu (read or write). Reads only.
func blockCompanionRuleLocked(c *Card) (companionRule, bool) {
	if c == nil || CatalogBlockRules == nil {
		return companionRule{}, false
	}
	key := catalogAbilityKeyOf(c)
	if key == "" {
		return companionRule{}, false
	}
	var out companionRule
	found := false
	for _, r := range CatalogBlockRules(key) {
		if !r.NotAlone {
			continue
		}
		cr := companionRule{need: max(r.NotAloneMinOthers, 1), greater: r.NotAloneOthersPowerGreater}
		if found {
			cr = cr.tighter(out)
		}
		out, found = cr, true
	}
	return out, found
}

// cantBlockAloneLocked reports whether `c` has a block companion rule
// at all (BlockRule.NotAlone).
//
// Caller must hold g.mu (read or write). Reads only.
func cantBlockAloneLocked(c *Card) bool {
	_, ok := blockCompanionRuleLocked(c)
	return ok
}

// CantBlockAlone reports whether `c` can't block alone (BlockRule.NotAlone,
// #5396): the view's cant_block_alone stamp, which the bot's block search
// reads to keep such a creature out of a block set it would be alone in.
// True for every companion rule (#6197), each of which at least forbids
// blocking alone. Reads the catalog only; safe under any lock the caller
// holds.
func CantBlockAlone(c *Card) bool { return cantBlockAloneLocked(c) }

// blockCompanionsOfLocked counts the creatures other than `b` that
// `b`'s controller blocks with in `assign` and that keep `b` company
// under `r` (CR 802.4b: another player's blocker is no company).
//
// Caller must hold g.mu (read or write). Reads only.
func (g *Game) blockCompanionsOfLocked(b *Card, r companionRule, assign blockAssignment) int {
	n := 0
	for id, atks := range assign {
		if len(atks) == 0 || id == b.InstanceID {
			continue
		}
		c := findBattlefieldCard(g, id)
		if c == nil || c.Controller != b.Controller {
			continue
		}
		if r.countsAsCompanion(b, c) {
			n++
		}
	}
	return n
}

// unmetBlockCompanionLocked is the first creature `seat` blocks with in
// `assign`, in battlefield order, whose companion rule `assign` does
// not meet, or nil: for "can't block alone", the one blocker of a
// one-blocker declaration.
//
// Caller must hold g.mu (read or write). Reads only.
func (g *Game) unmetBlockCompanionLocked(seat uuid.UUID, assign blockAssignment) *Card {
	if g.Battlefield == nil {
		return nil
	}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller != seat || len(assign[c.InstanceID]) == 0 {
			continue
		}
		r, ok := blockCompanionRuleLocked(c)
		if !ok {
			continue
		}
		if g.blockCompanionsOfLocked(c, r, assign) < r.need {
			return c
		}
	}
	return nil
}

// blockAloneRefusalLocked is the validator's companion check: for each
// defending player whose blockers `decls` names, the refusal when
// `after` leaves one of their blockers with fewer companions than its
// rule needs (alone, for "can't block alone"), or nil.
//
// Only the players the declaration touches are judged, so a short block
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
		short := g.unmetBlockCompanionLocked(b.Controller, after)
		if short == nil {
			continue
		}
		var attacker *Card
		if atks := after[short.InstanceID]; len(atks) > 0 {
			attacker = findBattlefieldCard(g, atks[0])
		}
		r, _ := blockCompanionRuleLocked(short)
		return g.blockRefusedErrorLocked(attacker, short,
			BlockRefusal{Reason: BlockReasonCantBlockAlone, Source: short.InstanceID, Label: blockCompanionLabel(r)})
	}
	return nil
}

// blockCompanionLabel is the refusal's clause for a companion rule
// other than "can't block alone" (#6197), or "" for that one.
func blockCompanionLabel(r companionRule) string {
	switch {
	case r.greater:
		return "a creature with greater power also blocks"
	case r.need > 1:
		return "at least " + countWord(r.need) + " other creatures block"
	}
	return ""
}

// takeOutLoneNotAloneBlockerLocked is the completion's half: while
// `seat`'s staged declaration has a creature whose companion rule it
// does not meet, that creature is taken out of combat before the
// declaration is announced (CR 509.1b: the declaration it was in is
// illegal). Taking one out can leave another short, so it repeats;
// each pass removes a blocker, so it ends. Nothing was announced for
// them yet, so nothing is taken back that anyone saw.
//
// Caller must hold g.mu in write mode.
func (g *Game) takeOutLoneNotAloneBlockerLocked(seat uuid.UUID) {
	if CatalogBlockRules == nil || g.Battlefield == nil {
		return
	}
	for {
		short := g.unmetBlockCompanionLocked(seat, g.currentBlockAssignmentLocked())
		if short == nil {
			return
		}
		g.removeFromCombatLocked(short)
	}
}

// notAloneBlockerGroupsLocked is the option generator's addition for a
// creature with a block companion rule, offered while `seat` has staged no
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
		r, ok := blockCompanionRuleLocked(b)
		if !ok {
			continue
		}
		if r.need > 1 {
			// #6197: a rule that needs several companions is offered as
			// one group of the first creatures that count, rather than
			// as every combination of them.
			if decls := g.blockCompanionGroupLocked(b, r, atk, pool, base, cache); decls != nil {
				out = append(out, decls)
				if len(out) >= limit {
					return out
				}
			}
			continue
		}
		for j, other := range pool {
			// Two creatures that each can't block alone pair once,
			// from the first of them.
			if j == i || (j < i && cantBlockAloneLocked(other)) {
				continue
			}
			if !r.countsAsCompanion(b, other) {
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

// blockCompanionGroupLocked is one declaration that meets `b`'s rule:
// `b` and the first r.need creatures of `pool` that count as its
// companions, all blocking `atk`, or nil when the pool cannot fill it
// or the validator refuses the group.
//
// Caller must hold g.mu with fresh layers. Reads only.
func (g *Game) blockCompanionGroupLocked(b *Card, r companionRule, atk *Card, pool []*Card, base blockAssignment, cache blockReachCache) []BlockDeclaration {
	decls := []BlockDeclaration{{Blocker: b.InstanceID, Attacker: atk.InstanceID}}
	for _, other := range pool {
		if len(decls) > r.need {
			break
		}
		if other.InstanceID == b.InstanceID || !r.countsAsCompanion(b, other) {
			continue
		}
		decls = append(decls, BlockDeclaration{Blocker: other.InstanceID, Attacker: atk.InstanceID})
	}
	if len(decls) <= r.need {
		return nil
	}
	if g.checkBlockDeclarationCachedLocked(base, decls, cache) != nil {
		return nil
	}
	return decls
}
