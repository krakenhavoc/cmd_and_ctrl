package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Criminal Past — Legendary Enchantment — Background {2}{B} (Commander
// Legends: Battle for Baldur's Gate, #5948):
//
//	"Commander creatures you own have menace and "This creature gets +X/+0,
//	 where X is the number of creature cards in your graveyard.""
//
// Menace is a layer-6 keyword grant to every commander creature the
// Background's controller owns, under anyone's control (Guild Artisan's
// recipient set). The pump is the second granted ability: a layer-7c static
// carried by a bundle (ADR 0093 Decision 10) that counts the creature cards
// in the graveyard of the creature's CONTROLLER, since "your" is whoever
// controls the commander. A stolen commander gives its new controller's
// graveyard. Choose a Background is deck construction (ADR 0144).
//
// No simplification.
const criminalPastGrant = "criminal-past/gets-plus-x-per-creature-card"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "54ad73d1-a17d-468c-a0ba-05c532a02cb2",
		Name:         "Criminal Past",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: criminalPastGrant,
			Static: []game.StaticAbility{{
				Label:    "This creature gets +X/+0, where X is the number of creature cards in your graveyard.",
				Layer:    game.Layer7PT,
				SubLayer: game.SubLayer7C_Modify,
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target != nil && source != nil && target.InstanceID == source.InstanceID
				},
				Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
					if p := g.PlayerByIDForEffect(source.Controller); p != nil && p.Graveyard != nil {
						for _, gc := range p.Graveyard.Cards {
							if gc.HasCardType("creature") {
								c.Power++
							}
						}
					}
				},
			}},
			Text: "This creature gets +X/+0, where X is the number of creature cards in your graveyard.",
		}},
		Static: []game.StaticAbility{
			effects.B16GrantKeywords(effects.CommanderCreatureYouOwn, "menace"),
			effects.GrantToCommanderCreaturesYouOwn(criminalPastGrant),
		},
	})
}
