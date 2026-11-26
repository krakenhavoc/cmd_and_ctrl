package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Predatory Urge — Enchantment — Aura {3}{G} (Planechase 2012, S82 OLD, #5569):
//
//	"Enchant creature
//	 Enchanted creature has "{T}: This creature deals damage equal to its power
//	 to target creature. That creature deals damage equal to its power to this
//	 creature.""
//
// A granted activated ability (Quicksilver Dagger's shape, ADR 0093): the
// enchanted creature's controller activates it and the creature is its source.
// The damage each way is the power as the ability resolves, dealt together as one
// damage event (CR 120.2), so deathtouch and lifelink on either side apply and a
// creature that has left deals none.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f6175cca-0e19-4db8-a175-3369fb650f24",
		Name:         "Predatory Urge",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Grants: []effects.AbilityGrant{{
			Key: predatoryUrgeGrant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{T}: This creature deals damage equal to its power to target creature. That creature deals damage equal to its power to this creature.",
				Purpose: game.Purpose{Answers: game.AnswerRemove},
				Cost:    effects.TapCost(),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.S82OldB432MutualDamage,
			}},
			Text: "{T}: This creature deals damage equal to its power to target creature. That creature deals damage equal to its power to this creature.",
		}},
		Static: []game.StaticAbility{effects.GrantAbilitiesToAttached(predatoryUrgeGrant)},
	})
}

// predatoryUrgeGrant names the granted ability's bundle: an on-disk identity.
const predatoryUrgeGrant = "predatory-urge/tap-mutual-damage"
