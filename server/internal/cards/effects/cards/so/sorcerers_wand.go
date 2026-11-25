package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sorcerer's Wand — Artifact — Equipment {1} (Dominaria, S82 OLD batch 214,
// #5530):
//
//	"Equipped creature has "{T}: This creature deals 1 damage to target player
//	 or planeswalker. If this creature is a Wizard, it deals 2 damage
//	 instead."
//	 Equip {3}"
//
// An ADR 0093 grant (Viridian Longbow's shape): the quoted ability is the
// equipped creature's own, so the damage is the creature's and the {T} cost
// keeps CR 302.6's summoning-sickness rule. "Is a Wizard" is read through the
// layers as the ability resolves, so a creature that was a Wizard when it was
// activated but is not now deals 1.
//
// No simplification.
func init() {
	const grant = "sorcerers-wand/ping"
	const text = "{T}: This creature deals 1 damage to target player or planeswalker. If this creature is a Wizard, it deals 2 damage instead."
	effects.Register(effects.Spec{
		OracleID:     "0edc8ad3-2920-4b25-8a86-98ce216111d1",
		Name:         "Sorcerer's Wand",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: grant,
			Activated: []effects.ActivatedAbility{{
				Label:   text,
				Cost:    effects.TapCost(),
				Targets: effects.TargetPlayerOrPlaneswalker(),
				Purpose: game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 2)), Answers: game.AnswerRemove},
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					amount := 1
					g.RecomputeLayersIfStaleLocked()
					if src, ok := g.LookupCardForEffect(item.SourceCardID); ok && src.HasSubtype(ctx.TextSubtype("Wizard")) {
						amount = 2
					}
					for _, t := range ctx.LegalTargets() {
						return effects.DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: amount}.Apply(ctx)
					}
					return nil
				},
			}},
			Text: text,
		}},
		Static:    []game.StaticAbility{effects.GrantAbilitiesToAttached(grant)},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{3}")},
	})
}
