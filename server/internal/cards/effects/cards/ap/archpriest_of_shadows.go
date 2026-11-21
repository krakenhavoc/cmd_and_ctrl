package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Archpriest of Shadows — Creature — Human Warlock {3}{B}{B}, 4/4 (March of
// the Machine, #5477):
//
//	"Backup 1 (When this creature enters, put a +1/+1 counter on target
//	 creature. If that's another creature, it gains the following abilities
//	 until end of turn.)
//	 Deathtouch
//	 Whenever this creature deals combat damage to a player or battle,
//	 return target creature card from your graveyard to the battlefield."
//
// Backup (CR 702.165) grants deathtouch and the combat-damage trigger, a
// bundle in Spec.Grants that the Archpriest also has itself. "Your
// graveyard" is that of the controller of the creature that has the
// ability. Damage to a battle is the same EventDealDamage as damage to a
// player (Bilbo, Unexpected Adventurer).
//
// No simplification.
const archpriestOfShadowsGrant = "archpriest-of-shadows/combat-damage-return"

func init() {
	effects.Register(effects.Spec{
		OracleID:        "7b03910a-2ca7-40cb-9c71-1c2cc856a287",
		Name:            "Archpriest of Shadows",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Grants: []effects.AbilityGrant{{
			Key:       archpriestOfShadowsGrant,
			Triggered: []game.TriggeredAbility{archpriestOfShadowsTrigger()},
			Text:      "Whenever this creature deals combat damage to a player or battle, return target creature card from your graveyard to the battlefield.",
		}},
		Triggered: []game.TriggeredAbility{
			effects.Backup("Archpriest of Shadows", 1, archpriestOfShadowsGrant, "deathtouch"),
			archpriestOfShadowsTrigger(),
		},
	})
}

func archpriestOfShadowsTrigger() game.TriggeredAbility {
	return effects.Targeting(
		effects.On(game.EventDealDamage, effects.ThisDealtCombatDamageToAPlayerOrBattle,
			"Archpriest of Shadows — return target creature card from your graveyard to the battlefield",
			effects.ReturnFirstLegalGraveyardTargetToBattlefield),
		effects.TargetCreatureInYourGraveyard())
}
