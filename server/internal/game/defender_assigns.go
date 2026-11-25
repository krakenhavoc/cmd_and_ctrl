package game

import "github.com/google/uuid"

// defender_assigns.go — who assigns an attacker's combat damage, and where
// an unblocked one may put it (#6192, ADR 0224 amendment 2026-11-25).
//
//   - Defensive Formation: "Rather than the attacking player, you assign
//     the combat damage of each creature attacking you. You can divide that
//     creature's combat damage as you choose among any of the creatures
//     blocking it." The card prints CR 702.22j's banding division for
//     every creature attacking its controller, so a blocked attacker's
//     division goes through the banding prompt, addressed to them
//     (queueBandingAssignmentPromptLocked), with the permanent named on
//     the frame (DefenderAssignsBy). An unblocked attacker, or one blocked
//     by a single creature, has nothing to divide (CR 510.1b-c).
//   - Cunning Giant: "If this creature is unblocked, you may have it
//     assign its combat damage to a creature defending player controls."
//     Its controller is asked, as the damage is assigned (CR 510.1), for
//     all of it to one such creature or all of it to what it attacks
//     (DamageAssignmentFrame.OneCreature).

// CatalogYouAssignAttackersDamage reports whether a battlefield permanent
// with this catalog key carries Defensive Formation's static. Populated
// from effects.Spec.YouAssignAttackersDamage. A nil hook grants nothing.
var CatalogYouAssignAttackersDamage func(key string) bool

// CatalogUnblockedAssignsToACreature reports whether a creature with this
// catalog key carries Cunning Giant's "If this creature is unblocked, you
// may have it assign its combat damage to a creature defending player
// controls." Populated from effects.Spec.UnblockedAssignsToACreature.
var CatalogUnblockedAssignsToACreature func(key string) bool

// defenderAssignsByLocked is the player who assigns `atk`'s combat damage
// instead of its controller, and the name of the permanent that says so:
// the player it is attacking, when they control a permanent with
// Defensive Formation's static (read live, so one that has lost its
// abilities grants nothing). uuid.Nil when nobody does. "Each creature
// attacking you": an attacker of that player's planeswalker or battle is
// not attacking them.
//
// Caller must hold g.mu.
func (g *Game) defenderAssignsByLocked(atk *Card) (uuid.UUID, string) {
	if atk == nil || CatalogYouAssignAttackersDamage == nil || g.Battlefield == nil {
		return uuid.Nil, ""
	}
	you := atk.AttackingTarget
	if you == uuid.Nil || g.classifyAttackTargetLocked(you) != AttackTargetPlayer {
		return uuid.Nil, ""
	}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller != you {
			continue
		}
		if key := catalogAbilityKeyOf(c); key != "" && CatalogYouAssignAttackersDamage(key) {
			return you, c.Name
		}
	}
	return uuid.Nil, ""
}

// unblockedCreatureTargetsLocked is, for an unblocked attacker with Cunning
// Giant's ability, the creatures the defending player controls that it may
// assign its damage to; nil otherwise.
//
// Caller must hold g.mu.
func (g *Game) unblockedCreatureTargetsLocked(atk *Card) []uuid.UUID {
	if atk == nil || CatalogUnblockedAssignsToACreature == nil {
		return nil
	}
	if key := catalogAbilityKeyOf(atk); key == "" || !CatalogUnblockedAssignsToACreature(key) {
		return nil
	}
	defender := g.defendingPlayerForAttackerLocked(atk)
	if defender == uuid.Nil {
		return nil
	}
	var out []uuid.UUID
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller == defender && c.IsCreature() {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// queueOneCreatureAssignmentPromptLocked asks the controller of the
// unblocked attacker `atk` whether it assigns its combat damage to one of
// `creatures` or, as it would anyway, to what it attacks (Cunning Giant,
// #6192).
//
// Caller must hold g.mu.
func (g *Game) queueOneCreatureAssignmentPromptLocked(atk *Card, creatures []uuid.UUID, atkPower int, step string) {
	frame := attackerDamageFrame(atk, creatures, atkPower, step)
	frame.AllowTrample = false
	frame.AsThoughUnblocked = true
	frame.OneCreature = true
	g.QueueChoiceForEffect(PendingChoice{
		Kind:             PendingChoiceDamageAssignment,
		Chooser:          atk.Controller,
		Count:            len(creatures),
		Source:           atk.InstanceID,
		Reason:           atk.Name + " is unblocked: you may have it assign its combat damage to a creature defending player controls",
		DamageAssignment: frame,
	})
}
