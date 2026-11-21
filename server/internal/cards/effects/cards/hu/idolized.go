package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Idolized — Enchantment — Aura {1}{W} (Fallout):
//
//	"Enchant creature
//	 Enchanted creature has "Whenever this creature attacks alone, it gets
//	 +X/+X until end of turn, where X is the number of nonland permanents
//	 you control.""
//
// Light of Promise's shape (a trigger the Aura grants the creature, so
// "you" is its controller). "Attacks alone" is the only creature declared
// as an attacker (CR 506.5, game.AttackedAlone), read when the trigger is
// harvested after the whole declaration. X is counted as the ability
// resolves, through the layers, so a nonland permanent that left in
// response is not counted and a land that became a creature is still a
// land.
//
// No simplification.
const idolizedGrant = "idolized/lone-attacker-pump"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "8cbf0173-2349-4205-b720-378643c347a4",
		Name:         "Idolized",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []effects.AbilityGrant{{
			Key: idolizedGrant,
			Triggered: []game.TriggeredAbility{
				effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return ev.CardID == source.InstanceID && game.AttackedAlone(g)
				}, "Idolized — this creature gets +X/+X until end of turn, where X is the number of nonland permanents you control",
					func(g *game.Game, item *game.StackItem) error {
						x := 0
						for _, c := range g.BattlefieldCardsForEffect() {
							if c.Controller == item.Controller && !c.IsLand() {
								x++
							}
						}
						return effects.BoostUntilEOT{Target: item.SourceCardID, Power: x, Toughness: x, Label: "Idolized — +X/+X"}.Apply(effects.NewContext(g, item))
					}),
			},
			Text: "Whenever this creature attacks alone, it gets +X/+X until end of turn, where X is the number of nonland permanents you control.",
		}},
		Static: []game.StaticAbility{effects.GrantAbilitiesToAttached(idolizedGrant)},
	})
}
