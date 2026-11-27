package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pillory of the Sleepless — Enchantment — Aura {1}{W}{B} (Guildpact):
//
//	"Enchant creature
//	 Enchanted creature can't attack or block.
//	 Enchanted creature has "At the beginning of your upkeep, you lose 1
//	 life.""
//
// Arrest's restriction, and Inevitable End's granted upkeep trigger: "you"
// is the enchanted creature's controller, who loses the life.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "74776491-29c3-46e6-989b-cf9a00bbfcef",
		Name:         "Pillory of the Sleepless",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Grants: []effects.AbilityGrant{{
			Key: pilloryGrant,
			Triggered: []game.TriggeredAbility{
				effects.AtYourUpkeep("Pillory of the Sleepless — you lose 1 life", effects.Do(effects.YouLoseLife(1))),
			},
			Text: "At the beginning of your upkeep, you lose 1 life.",
		}},
		Static: []game.StaticAbility{
			effects.RestrictAttached(game.CantAttackOrBlock),
			effects.GrantAbilitiesToAttached(pilloryGrant),
		},
	})
}

// pilloryGrant names the granted trigger's bundle: an on-disk identity.
const pilloryGrant = "pillory-of-the-sleepless/upkeep-lose-1-life"
