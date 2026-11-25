package game

import "github.com/google/uuid"

// attack_requirements_player.go is CR 508.1d's player-level requirement
// (#6087): "Each opponent must attack with at least one creature each
// combat if able" (Seeker of Slaanesh) and "Each opponent must attack
// you or a planeswalker you control with at least one creature each
// combat if able" (Trove of Temptation).
//
// # What it is
//
// The requirement binds the PLAYER, not a creature: no one creature of
// theirs has to attack, but some one of them does when one can. Making
// every creature "attack if able" would be stronger than printed. So it
// is ONE requirement for CR 508.1d's count — the declaration obeys it
// when at least one of the player's attacking creatures attacks a target
// its shape accepts, and a second creature attacking the same way obeys
// nothing more. Two Seekers under two opponents are two requirements,
// and one attacker obeys both.
//
// # Where it rides
//
// On the creatures, like every other requirement: an ordinary layer
// static writes it onto each creature the source's opponents control
// (AttackRequirement.AtLeastOneCreature), so the layer pass decides who
// is bound and the source's departure ends it. The solver reads it off
// them as the player's: attackGroupsLocked collects the distinct ones
// across the declaring player's creatures, counts each obeyed by the
// staged declaration once, and hands the rest to the search as a bonus
// (groupBonus) that the first obeying attack collects.
//
// # The search
//
// maxRequirementAssignmentWithGroups carries the bonus as a concave
// reward on the min-cost flow: a cap-1 arc that pays it beside a free
// arc, on the combat-wide node for "any attack" and on a node per player
// for "that player or a planeswalker they control". The two families are
// laminar (a player's targets are a subset of every target), so one
// network holds them and successive shortest paths stay exact. Every
// restriction and capacity the creature requirements already honour
// (attack limits, taxes — a pair that costs something is never offered,
// CR 508.1d's cost clause — "can't attack", defenders, directions)
// applies unchanged, because the bonus only changes what a pair is
// worth, never which pairs exist.

// groupBonus is the player-level requirements a declaration does not
// obey yet: how many any attack would obey, and how many an attack on
// each player (or a planeswalker they control) would obey. playerOf maps
// each attack target the declaring player has to the player whose
// requirements it obeys.
type groupBonus struct {
	any       int
	perPlayer map[uuid.UUID]int
	playerOf  map[uuid.UUID]uuid.UUID
}

func (gb groupBonus) empty() bool { return gb.any == 0 && len(gb.perPlayer) == 0 }

// gain is what the first attack at `target` would add.
func (gb groupBonus) gain(target uuid.UUID) int {
	n := gb.any
	if p, ok := gb.playerOf[target]; ok {
		n += gb.perPlayer[p]
	}
	return n
}

// attackGroupsLocked returns how many of the declaring player's
// player-level requirements `assign` obeys, and the rest as a bonus for
// the search.
//
// Caller must hold g.mu with fresh layers. Reads only.
func (g *Game) attackGroupsLocked(ap uuid.UUID, assign map[uuid.UUID]uuid.UUID) (int, groupBonus) {
	var gb groupBonus
	groups := g.playerAttackGroupsLocked(ap, assign)
	if len(groups) == 0 {
		return 0, gb
	}
	obeyed := 0
	for _, r := range groups {
		if g.groupObeyedInLocked(r, ap, assign) {
			obeyed++
			continue
		}
		if p := r.AttacksPlayerOrTheirPlaneswalkers; p != uuid.Nil {
			if gb.perPlayer == nil {
				gb.perPlayer = map[uuid.UUID]int{}
			}
			gb.perPlayer[p]++
			continue
		}
		gb.any++
	}
	if len(gb.perPlayer) > 0 {
		gb.playerOf = map[uuid.UUID]uuid.UUID{}
		for _, t := range g.AttackTargetsForEffect(ap) {
			for p := range gb.perPlayer {
				probe := AttackRequirement{AttacksPlayerOrTheirPlaneswalkers: p}
				if probe.obeyedBy(g, t.ID) {
					gb.playerOf[t.ID] = p
					break
				}
			}
		}
	}
	return obeyed, gb
}

// playerAttackGroupsLocked is every distinct player-level requirement on
// a creature `ap` controls, in battlefield order. Identical requirements
// from one source on many creatures are one requirement.
//
// Caller must hold g.mu with fresh layers.
func (g *Game) playerAttackGroupsLocked(ap uuid.UUID, assign map[uuid.UUID]uuid.UUID) []AttackRequirement {
	if ap == uuid.Nil || g.Battlefield == nil {
		return nil
	}
	var out []AttackRequirement
	var seen map[AttackRequirement]bool
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller != ap || !c.IsCreature() {
			continue
		}
		for _, r := range g.attackRequirementsOfLocked(c, assign) {
			if !r.AtLeastOneCreature || seen[r] {
				continue
			}
			if seen == nil {
				seen = map[AttackRequirement]bool{}
			}
			seen[r] = true
			out = append(out, r)
		}
	}
	return out
}

// groupObeyedInLocked reports whether one of `ap`'s attacking creatures
// in `assign` attacks a target the player-level requirement `r` accepts.
//
// Caller must hold g.mu.
func (g *Game) groupObeyedInLocked(r AttackRequirement, ap uuid.UUID, assign map[uuid.UUID]uuid.UUID) bool {
	for id, target := range assign {
		c := findBattlefieldCard(g, id)
		if c == nil || c.Controller != ap {
			continue
		}
		if r.obeyedBy(g, target) {
			return true
		}
	}
	return false
}

// withoutObeyedGroupsLocked drops from `reqs` every player-level
// requirement the active player's declaration `assign` already obeys,
// so a refusal never names one that is satisfied.
//
// Caller must hold g.mu.
func (g *Game) withoutObeyedGroupsLocked(reqs []AttackRequirement, assign map[uuid.UUID]uuid.UUID) []AttackRequirement {
	ap := g.activePlayerIDLocked()
	out := make([]AttackRequirement, 0, len(reqs))
	for _, r := range reqs {
		if r.AtLeastOneCreature && g.groupObeyedInLocked(r, ap, assign) {
			continue
		}
		out = append(out, r)
	}
	return out
}
