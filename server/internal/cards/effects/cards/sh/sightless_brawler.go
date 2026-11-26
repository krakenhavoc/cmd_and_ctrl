package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const sightlessBrawlerGrant = "sightless-brawler/cant-attack-alone"

// Sightless Brawler — Enchantment Creature — Human Warrior {1}{W}, 3/2 (Journey into Nyx):
//
//	"Bestow {4}{W}
//	 This creature can't attack alone.
//	 Enchanted creature gets +3/+2 and can't attack alone."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0f7d76a9-3f47-4cfa-882b-1790256d15b2",
		Name:         "Sightless Brawler",
		Completeness: effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{
			effects.Bestow("{4}{W}"),
		},
		AttackLimits: []game.AttackLimit{effects.CantAttackAlone()},
		Grants: []effects.AbilityGrant{{
			Key:          sightlessBrawlerGrant,
			AttackLimits: []game.AttackLimit{effects.CantAttackAlone()},
			Text:         "This creature can't attack alone.",
		}},
		Static: []game.StaticAbility{
			effects.PumpAttached(3, 2),
			effects.GrantAbilitiesToAttached(sightlessBrawlerGrant),
		},
	})
}
