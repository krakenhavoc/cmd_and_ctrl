package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2_b03_helpers.go — card-side helpers for Commander-deck wave 2,
// batch 3 (#5368).

// Token rows added from this file's init so the shared table's column
// alignment is not rewritten.
func init() {
	for key, card := range map[string]game.Card{
		// Sandwurm Convergence (the printed Wurm of Duskmourn Commander).
		"5/5 green Wurm": {Name: "Wurm", TypeLine: "Token Creature — Wurm", Power: 5, Toughness: 5, Colors: []string{"G"}},
	} {
		if _, dup := tokenTable[key]; dup {
			panic("effects: duplicate token row " + key)
		}
		tokenTable[key] = card
	}
}

// cmdr2B03CreaturesMatchingCantAttackYouOrPlaneswalkers is "<creatures
// matching the queries> can't attack you or planeswalkers you control"
// (CR 508.1c): CreaturesMatchingCantAttackYou's wider sibling, covering
// the controller's planeswalkers (and battles) as well as the player.
func cmdr2B03CreaturesMatchingCantAttackYouOrPlaneswalkers(label string, attackers ...game.PermanentQuery) game.StaticAbility {
	return game.StaticAbility{
		Label: label,
		Layer: game.Layer6Ability,
		AppliesTo: func(target *game.Card, _ *game.Game, _ *game.Card) bool {
			return target.IsCreature()
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
			if source == nil {
				return
			}
			c.AttackTargetRestrictions = append(c.AttackTargetRestrictions, game.AttackTargetRestriction{
				Source:                source.InstanceID,
				SourceName:            source.Name,
				NotPlayer:             source.Controller,
				OnlyAttackersMatching: append([]game.PermanentQuery(nil), attackers...),
			})
		},
	}
}

// cmdr2B03AttackingYouOrYourPlaneswalker is "creature that's attacking you
// or a planeswalker you control" (Soul Snare): AttackingYou's wider
// sibling. The attacked object is the caster as a player, or a
// planeswalker on the battlefield the caster controls (CR 506.3).
func cmdr2B03AttackingYouOrYourPlaneswalker() CardPredicate {
	return func(g *game.Game, caster uuid.UUID, c game.Card) bool {
		return attacksYouOrYourPlaneswalker(g, c.AttackingTarget, caster)
	}
}

// cmdr2B03DuringSourcesControllersTurn passes while it is the turn of the
// modifier source's controller: "during your turn" on a static ability.
func cmdr2B03DuringSourcesControllersTurn() CostPredicate {
	return func(q game.CostQuery) bool {
		return q.Game != nil && IsYourTurn(q.Game, q.Source.Controller)
	}
}

// cmdr2B03NotAManaAbility passes on an activation that is not a mana
// ability (CR 605.1a): "unless they're mana abilities".
func cmdr2B03NotAManaAbility() CostPredicate {
	return func(q game.CostQuery) bool {
		return q.Ability != nil && !q.Ability.Mana
	}
}
