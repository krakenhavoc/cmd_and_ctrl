package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap2_b04_b1_helpers.go — the card side of the player-level CR
// 508.1d requirement (#6087; the engine half is
// game/attack_requirements_player.go).
//
//	Static: []game.StaticAbility{EachOpponentAttacksWithAtLeastOneCreature()},           // Seeker of Slaanesh
//	Static: []game.StaticAbility{EachOpponentAttacksYouWithAtLeastOneCreature()},        // Trove of Temptation
//
// The requirement binds each opponent, not a creature: it is written
// onto every creature an opponent of the source's controller controls,
// as game.AttackRequirement.AtLeastOneCreature, and the engine counts it
// once per player however many creatures carry it. Like
// AttacksEachCombatWhere it is the source's ability: a source that loses
// its abilities or leaves stops imposing it.

// atLeastOneCreatureStatic is the shared static; `you` makes it "attack
// you or a planeswalker you control".
func atLeastOneCreatureStatic(label string, you bool) game.StaticAbility {
	return game.StaticAbility{
		Layer: game.Layer6Ability,
		Label: label,
		AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return target.IsCreature() && target.Controller != source.Controller
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
			r := requirementFrom(source, uuid.Nil)
			r.AtLeastOneCreature = true
			if you && source != nil {
				r.AttacksPlayerOrTheirPlaneswalkers = source.Controller
			}
			c.AttackRequirements = append(c.AttackRequirements, r)
		},
	}
}

// EachOpponentAttacksWithAtLeastOneCreature is "Each opponent must
// attack with at least one creature each combat if able" (Seeker of
// Slaanesh's Allure of Slaanesh).
func EachOpponentAttacksWithAtLeastOneCreature() game.StaticAbility {
	return atLeastOneCreatureStatic("Each opponent must attack with at least one creature each combat if able.", false)
}

// EachOpponentAttacksYouWithAtLeastOneCreature is "Each opponent must
// attack you or a planeswalker you control with at least one creature
// each combat if able" (Trove of Temptation): obeyed only by an attack
// on the source's controller or a planeswalker they control.
func EachOpponentAttacksYouWithAtLeastOneCreature() game.StaticAbility {
	return atLeastOneCreatureStatic("Each opponent must attack you or a planeswalker you control with at least one creature each combat if able.", true)
}
