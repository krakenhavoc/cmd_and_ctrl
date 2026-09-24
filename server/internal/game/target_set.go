package game

import (
	"fmt"

	"github.com/google/uuid"
)

// target_set.go — #1559: legality of the chosen SET of targets
// (CR 601.2c), and the "X or less" bound that rides with it.
//
// Everything else about a target clause judges ONE candidate at a
// time: TargetSpec.CardOK, the zone list, the CR 702 keyword gate.
// Some printed clauses constrain the picks against each other —
// "any number of target creature cards that each have a different
// mana value" (Agadeem's Awakening), "two target creatures controlled
// by different players" (Run Away Together) — and no per-candidate
// predicate can say that, because whether a pick is legal depends on
// what else was picked.
//
// The rule is DECLARATIVE: a key read off each pick, which no two
// picks of the clause may share. Four readers answer one question
// with it, and a key is the shape all four can evaluate:
//
//   - the announce gate (validateAnnouncedTargetsWithLocked), which
//     refuses a set that breaks it with a sentence naming the rule;
//   - the CR 608.2b re-check (TargetStillLegalForEffect), which judges
//     the SURVIVING picks against it — see setRuleConflictLocked;
//   - the enumerator (internal/legal), which never builds a set that
//     breaks it;
//   - the client's picker, which greys a candidate whose key an
//     earlier pick already holds, from the map the view ships.
//
// See docs/decisions/0019-structured-targeting.md, the 2026-09-24
// amendment.

// TargetDifference is a clause's set rule: no two picks may share the
// value Key reads.
type TargetDifference struct {
	// Label is the printed rule in a form the picker and the refusal
	// can both say after "the targets must": "each have a different
	// mana value", "be controlled by different players".
	Label string

	// Key reads the value that must differ, off a card in the zone it
	// was found in. ok=false means the card has no value to compare
	// (an unreadable cost, a card the rule does not apply to), and a
	// pick with no key conflicts with nothing — the lenient reading,
	// chosen because a key the engine cannot compute is one the
	// client cannot grey either.
	//
	// Player picks have no key: no printed clause of this family
	// constrains a set of players.
	//
	// Runs under g.mu. MUST NOT call public locking mutators.
	Key func(g *Game, c Card, zone ZoneKind) (key string, ok bool)
}

// xBoundAdmits applies ManaValueAtMostX to a candidate card: true
// when the clause has no bound, or the bound is not yet known (a hand
// snapshot, built before X is announced), or the card's mana value is
// at most X. A card whose cost cannot be read does not meet a bound
// it might not meet — Card.ParsedManaValue's rule.
func (s *TargetSpec) xBoundAdmits(c Card) bool {
	if s == nil || !s.ManaValueAtMostX || !s.xBoundSet {
		return true
	}
	mv, ok := c.ParsedManaValue()
	return ok && mv <= s.xBound
}

// bindStepsX writes the announced X onto every X-bounded step of an
// announcement. The steps hold clause COPIES (AnnouncedClauses), so
// the catalog's shared declaration is never touched — the same
// reasoning resolveStepCountsFromX gives for CountFromX.
func bindStepsX(steps []AnnouncedClause, x int) {
	for i := range steps {
		if steps[i].Clause.ManaValueAtMostX {
			steps[i].Clause.xBound, steps[i].Clause.xBoundSet = x, true
		}
	}
}

// BindStepsXForEffect is bindStepsX for the bot's enumerator, which
// builds an announcement's steps itself and has to judge its target
// sets under the X it is about to announce, exactly as the engine
// will. Pure; takes no lock.
func BindStepsXForEffect(steps []AnnouncedClause, x int) {
	bindStepsX(steps, x)
}

// StepsBoundByX reports whether any step's legality depends on the
// announced X through ManaValueAtMostX. Pure.
func StepsBoundByX(steps []AnnouncedClause) bool {
	for i := range steps {
		if steps[i].Clause.ManaValueAtMostX {
			return true
		}
	}
	return false
}

// targetDifferenceKeyLocked is the set-rule key of one pick under
// clause, or false when the clause has no rule or the pick has no
// key: a player, a card that is gone, a card the Key declines.
//
// Caller must hold g.mu.
func (g *Game) targetDifferenceKeyLocked(clause *TargetClause, ref TargetRef) (string, bool) {
	if clause == nil || clause.Different == nil || clause.Different.Key == nil || ref.Kind != TargetCard {
		return "", false
	}
	z := g.findCardZoneLocked(ref.ID)
	if z == nil {
		return "", false
	}
	for i := range z.Cards {
		if z.Cards[i].InstanceID == ref.ID {
			return clause.Different.Key(g, z.Cards[i], z.Kind)
		}
	}
	return "", false
}

// TargetDifferenceKeysForEffect is the set-rule key of each card in
// ids under spec, for the protocol projection and the enumerator. Nil
// when the spec has no rule; a card with no key is absent from the
// map. Caller must hold g.mu.
func (g *Game) TargetDifferenceKeysForEffect(spec *TargetSpec, ids []uuid.UUID) map[uuid.UUID]string {
	if spec == nil || spec.Different == nil {
		return nil
	}
	out := make(map[uuid.UUID]string, len(ids))
	for _, id := range ids {
		if k, ok := g.targetDifferenceKeyLocked(spec, TargetRef{Kind: TargetCard, ID: id}); ok {
			out[id] = k
		}
	}
	return out
}

// ManaValuesForEffect is each card's mana value, for the view to ship
// beside an X-bounded clause's legal set so the client can narrow it
// by the X it collected. A card whose cost cannot be read is absent —
// it meets no bound (xBoundAdmits). Caller must hold g.mu.
func (g *Game) ManaValuesForEffect(ids []uuid.UUID) map[uuid.UUID]int {
	out := make(map[uuid.UUID]int, len(ids))
	for _, id := range ids {
		c := g.findCardByIDLocked(id)
		if c == nil {
			continue
		}
		if mv, ok := c.ParsedManaValue(); ok {
			out[id] = mv
		}
	}
	return out
}

// targetSetError is the announce gate's refusal for a set that breaks
// its clause's rule. It wraps ErrIllegalTarget, so every caller that
// already maps an illegal target keeps working, and it carries the
// printed rule, so the toast says which rule was broken rather than
// "illegal target" about a set whose members are each fine.
func targetSetError(d *TargetDifference) error {
	label := "each be different"
	if d != nil && d.Label != "" {
		label = d.Label
	}
	return fmt.Errorf("%w: those targets must %s", ErrIllegalTarget, label)
}

// setRuleConflictLocked is the CR 608.2b half of the set rule: does
// ref, which has already passed its own clause's per-candidate
// re-check, share its key with another SURVIVING pick of the same
// clause?
//
// The choice this makes, stated because CR 608.2b says a target must
// still meet the targeting requirements without saying how a
// requirement BETWEEN targets is apportioned: the rule is judged over
// the picks that are still individually legal, and when two of them
// now share a key, BOTH are illegal. Neither is preferred — the
// constraint is symmetric, and "the first one you named" is not a
// rule of the game — and it is the WEAKER of the available readings,
// never the stronger: a Run Away Together whose two creatures end up
// under one controller returns neither. A pick that became illegal on
// its own is dropped first and conflicts with nothing, so one creature
// card leaving in response to Agadeem's Awakening costs exactly that
// card.
//
// Caller must hold g.mu.
func (g *Game) setRuleConflictLocked(item *StackItem, clause *TargetClause, ref TargetRef) bool {
	key, ok := g.targetDifferenceKeyLocked(clause, ref)
	if !ok {
		return false
	}
	src := g.stackItemSourceLocked(item)
	for _, other := range item.Targets {
		if other.Kind != TargetCard || other.ID == ref.ID {
			continue
		}
		if other.Mode != ref.Mode || other.Slot != ref.Slot {
			continue
		}
		if !g.targetLegalLocked(src, clause, other) {
			continue
		}
		if k, ok := g.targetDifferenceKeyLocked(clause, other); ok && k == key {
			return true
		}
	}
	return false
}

// fillableCountLocked is the most picks one announcement could make
// from lt under clause: every player and card, except that cards
// sharing a set-rule key count once (#1559). "Two target creatures
// controlled by different players" with every creature on one side
// of the table has two legal candidates and no legal pair, and a
// CR 603.3d / cast-offer check that counted candidates would open a
// prompt nobody can answer.
//
// Caller must hold g.mu.
func (g *Game) fillableCountLocked(clause *TargetClause, lt LegalTargets) int {
	n := len(lt.Players)
	if clause == nil || clause.Different == nil {
		return n + len(lt.Cards)
	}
	seen := make(map[string]bool, len(lt.Cards))
	for _, id := range lt.Cards {
		k, ok := g.targetDifferenceKeyLocked(clause, TargetRef{Kind: TargetCard, ID: id})
		if !ok {
			n++
			continue
		}
		if !seen[k] {
			seen[k] = true
			n++
		}
	}
	return n
}

// PickTargetSetRuleForEffect is the set rule of the clause a
// pick_target prompt is asking about, with the key of each candidate
// it offers, for the wire projection and the enumerator (#1559). Nil
// rule when the clause has none, or when the prompt carries no live
// frame (a restored snapshot), in which case the announce gate still
// enforces the rule on the answer.
//
// Caller must hold g.mu.
func (g *Game) PickTargetSetRuleForEffect(c *PendingChoice) (*TargetDifference, map[uuid.UUID]string) {
	if c == nil || c.Kind != PendingChoicePickTarget || c.pickTargetResume == nil {
		return nil, nil
	}
	clause := c.pickTargetResume.currentClause()
	if clause == nil || clause.Different == nil {
		return nil, nil
	}
	return clause.Different, g.TargetDifferenceKeysForEffect(clause, c.PickTargetCards)
}

// withoutSetRuleConflictsLocked narrows slot `slot`'s retarget
// alternatives to the cards whose set-rule key no OTHER slot of the
// same clause holds (#1559). CR 115.7c: the new target must not make
// the unchanged ones illegal, and under a set rule a shared key is
// exactly that. The gate (retargetCheckLocked) refuses such a change
// anyway; this keeps it off the offer, so a prompt never lists an
// answer the engine would refuse.
//
// Caller must hold g.mu.
func (g *Game) withoutSetRuleConflictsLocked(clause *TargetClause, lt LegalTargets, targets []TargetRef, slot int) LegalTargets {
	if clause == nil || clause.Different == nil || slot < 0 || slot >= len(targets) {
		return lt
	}
	held := make(map[string]bool, len(targets))
	for i, t := range targets {
		if i == slot || t.Mode != targets[slot].Mode || t.Slot != targets[slot].Slot {
			continue
		}
		if k, ok := g.targetDifferenceKeyLocked(clause, t); ok {
			held[k] = true
		}
	}
	if len(held) == 0 {
		return lt
	}
	out := LegalTargets{Players: lt.Players}
	for _, id := range lt.Cards {
		if k, ok := g.targetDifferenceKeyLocked(clause, TargetRef{Kind: TargetCard, ID: id}); ok && held[k] {
			continue
		}
		out.Cards = append(out.Cards, id)
	}
	return out
}

// TargetsWithinXForEffect reports whether every pick that answers an
// X-bounded step ("with mana value X or less") meets that bound at x
// (#1559) — the enumerator's re-check for a set it priced at an X
// other than the one it bound the steps to. Picks of other steps pass.
//
// Caller must hold g.mu.
func (g *Game) TargetsWithinXForEffect(steps []AnnouncedClause, targets []TargetRef, x int) bool {
	for _, t := range targets {
		if t.Kind != TargetCard {
			continue
		}
		for i := range steps {
			if steps[i].Mode != t.Mode || steps[i].Slot != t.Slot || !steps[i].Clause.ManaValueAtMostX {
				continue
			}
			c := g.findCardByIDLocked(t.ID)
			if c == nil {
				return false
			}
			if mv, ok := c.ParsedManaValue(); !ok || mv > x {
				return false
			}
		}
	}
	return true
}
