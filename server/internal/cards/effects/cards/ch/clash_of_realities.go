package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Clash of Realities — Enchantment {3}{R} (Betrayers of Kamigawa, #5602):
//
//	"All Spirits have "When this permanent enters, you may have it deal 3
//	 damage to target non-Spirit creature."
//	 Non-Spirit creatures have "When this creature enters, you may have it
//	 deal 3 damage to target Spirit creature.""
//
// Two granted enters triggers (ADR 0093, CR 603.6a), Harmonic Sliver's
// shape, one bundle per recipient set: every Spirit at the table, and every
// creature that is not a Spirit. Each is an optional, targeted trigger whose
// "it" is the entering permanent, so the damage comes from it; a changeling
// is a Spirit and gets only the first. Spirit is a word of the
// recipient's text (CR 612.2) and is read through the layered view.
//
// No simplification.
const (
	clashOfRealitiesSpiritGrant = "clash-of-realities/spirit"
	clashOfRealitiesOtherGrant  = "clash-of-realities/non-spirit"
)

func init() {
	spiritLabel := "Clash of Realities — you may have it deal 3 damage to target non-Spirit creature"
	otherLabel := "Clash of Realities — you may have it deal 3 damage to target Spirit creature"
	effects.Register(effects.Spec{
		OracleID:     "7ab18e48-f935-405b-8658-b49052d20529",
		Name:         "Clash of Realities",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{
			{
				Key: clashOfRealitiesSpiritGrant,
				Triggered: []game.TriggeredAbility{
					effects.Optional(effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisEnters(spiritLabel, effects.SourceDealsDamageToEachLegalTarget(3)),
						effects.TargetCreature("target non-Spirit creature", effects.Not(effects.HasSubtype("Spirit")))),
						effects.ForTargets(effects.DamageToTarget(0, 3))), "Clash of Realities — have it deal 3 damage to the target?"),
				},
				Text: "When this permanent enters, you may have it deal 3 damage to target non-Spirit creature.",
			},
			{
				Key: clashOfRealitiesOtherGrant,
				Triggered: []game.TriggeredAbility{
					effects.Optional(effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisEnters(otherLabel, effects.SourceDealsDamageToEachLegalTarget(3)),
						effects.TargetCreature("target Spirit creature", effects.HasSubtype("Spirit"))),
						effects.ForTargets(effects.DamageToTarget(0, 3))), "Clash of Realities — have it deal 3 damage to the target?"),
				},
				Text: "When this creature enters, you may have it deal 3 damage to target Spirit creature.",
			},
		},
		Static: []game.StaticAbility{
			effects.GrantAbilities(func(target *game.Card, _ *game.Game, _ *game.Card) bool {
				return target.HasSubtypeWord("Spirit")
			}, clashOfRealitiesSpiritGrant),
			effects.GrantAbilities(func(target *game.Card, _ *game.Game, _ *game.Card) bool {
				return target.IsCreature() && !target.HasSubtypeWord("Spirit")
			}, clashOfRealitiesOtherGrant),
		},
	})
}
