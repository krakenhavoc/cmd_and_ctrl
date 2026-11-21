package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bear Umbra — Enchantment — Aura {2}{G}{G} (EDHREC rank 899, #3654):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+2 and has "Whenever this creature
//	 attacks, untap all lands you control."
//	 Umbra armor (If enchanted creature would be destroyed, instead
//	 remove all damage from it and destroy this Aura.)"
//
// +2/+2 is a layer 7c pump on the enchanted creature, the quoted attack
// trigger is a granted bundle (ADR 0093), so "you" is the creature's
// controller, and umbra armor is the CR 702.89 replacement
// (UmbraArmor, game/umbra_armor.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     BearUmbraOracleID,
		Name:         "Bear Umbra",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []effects.AbilityGrant{{
			Key: bearUmbraUntap,
			Triggered: []game.TriggeredAbility{{
				Watches:   []game.EventKind{game.EventAttack},
				AppliesTo: effects.ThisAttacked,
				Key:       "Bear Umbra — untap all lands you control",
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.UntapAllLandsControlledBy(g, item.Controller, effects.NewContext(g, item))
				},
			}},
			Text: "Whenever this creature attacks, untap all lands you control.",
		}},
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 2),
			effects.GrantAbilitiesToAttached(bearUmbraUntap),
		},
		Replacements: []game.ReplacementEffect{effects.UmbraArmor()},
	})
}

const (
	BearUmbraOracleID = "44661108-d72f-4d02-9303-7e4b0ecf7e9c"
	bearUmbraUntap    = "bear-umbra/untap"
)
