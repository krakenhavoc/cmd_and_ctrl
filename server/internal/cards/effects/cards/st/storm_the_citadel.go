package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Storm the Citadel — Sorcery {4}{G} (War of the Spark, #5520):
//
//	"Until end of turn, creatures you control get +2/+2 and gain "Whenever
//	 this creature deals combat damage to a player or planeswalker, destroy
//	 target artifact or enchantment defending player controls.""
//
// Furnace Reins's granted trigger (ADR 0093) on the player-or-planeswalker
// combat damage predicate Grateful Apparition uses, its target chosen from
// the artifacts and enchantments the defending player controls as the trigger
// goes on the stack (Floral Spuzzem's clause). The +2/+2 and the grant are one
// effect over the creatures you control as the spell resolves (CR 611.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bd4810a4-8f86-4d1c-82f1-cbcd7e9479d6",
		Name:         "Storm the Citadel",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: stormTheCitadelGrant,
			Triggered: []game.TriggeredAbility{
				effects.WithTargetsFrom(
					effects.On(game.EventDealDamage, effects.ThisDealtCombatDamageToAPlayerOrPlaneswalker,
						"Storm the Citadel — destroy target artifact or enchantment defending player controls",
						effects.DestroyFirstLegalCardTarget),
					effects.TargetOfDefendingPlayer("permanent", "target artifact or enchantment defending player controls",
						effects.Or(effects.Artifact(), effects.Enchantment()))),
			},
			Text: "Whenever this creature deals combat damage to a player or planeswalker, destroy target artifact or enchantment defending player controls.",
		}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.GrantAbilitiesFor{
				Match: effects.And(effects.Creature(), effects.YouControl()),
				Keys:  []string{stormTheCitadelGrant},
				Also:  []game.Mod{game.ModifyPTMod(2, 2)},
				Label: "Storm the Citadel — +2/+2 and a combat damage trigger until end of turn",
			}.Apply(ctx)
		},
	})
}

// stormTheCitadelGrant is the granted bundle's key. An on-disk identity: never
// renamed, never reused.
const stormTheCitadelGrant = "storm-the-citadel/combat-damage-destroy"
