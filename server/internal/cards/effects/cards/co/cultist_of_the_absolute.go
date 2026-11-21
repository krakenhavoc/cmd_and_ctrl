package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cultist of the Absolute — Legendary Enchantment — Background {B} (Commander
// Legends: Battle for Baldur's Gate, #5948):
//
//	"Commander creatures you own get +3/+3 and have flying, deathtouch,
//	 "Ward—Pay 3 life," and "At the beginning of your upkeep, sacrifice a
//	 creature.""
//
// The anthem and the two evergreen keywords are statics on each commander
// creature the Background's controller owns, under anyone's control (Raised
// by Giants' recipient set): +3/+3 in layer 7c, flying and deathtouch in layer
// 6. Ward and the upkeep trigger are granted abilities (ADR 0093): the
// commander's own ward (Purging Stormbrood's WardLife, charged to the player
// who targeted it) and its own upkeep trigger, so "your" is whoever controls
// the commander and the creature sacrificed may be the commander itself.
// Choose a Background is deck construction (ADR 0144).
//
// No simplification.
const cultistOfTheAbsoluteGrant = "cultist-of-the-absolute/ward-and-sacrifice"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "bdc2a72a-f1c0-4e67-8e86-1121a4000537",
		Name:         "Cultist of the Absolute",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: cultistOfTheAbsoluteGrant,
			Triggered: []game.TriggeredAbility{
				effects.Ward(effects.WardLife(3), "Cultist of the Absolute — ward, pay 3 life"),
				effects.AtYourUpkeep("Cultist of the Absolute — sacrifice a creature",
					func(g *game.Game, item *game.StackItem) error {
						g.PlayerSacrificesForEffect(item.SourceCardID, item.Controller,
							effects.SacrificeSpec("a creature", effects.Creature()),
							"Cultist of the Absolute — sacrifice a creature")
						return nil
					}),
			},
			Text: "Ward—Pay 3 life. At the beginning of your upkeep, sacrifice a creature.",
		}},
		Static: []game.StaticAbility{
			effects.B16Anthem(effects.CommanderCreatureYouOwn, 3, 3),
			effects.B16GrantKeywords(effects.CommanderCreatureYouOwn, "flying", "deathtouch"),
			effects.GrantToCommanderCreaturesYouOwn(cultistOfTheAbsoluteGrant),
		},
	})
}
