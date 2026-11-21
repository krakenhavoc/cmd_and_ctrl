package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Combat Research — Enchantment — Aura {U} (Dominaria United):
//
//	"Enchant creature
//	 Enchanted creature has "Whenever this creature deals combat damage
//	 to a player, draw a card."
//	 As long as enchanted creature is legendary, it gets +1/+1 and has
//	 ward {1}."
//
// The draw is a granted bundle (ADR 0093), the same trigger Catlike
// Curiosity grants, so its controller draws. The legendary clause is read
// live through the layered view on every recompute and every time the
// ward would trigger: the pump (layer 7c) and the ward both switch on and
// off with the creature's supertype.
//
// No simplification.
func init() {
	legendary := func(target *game.Card, g *game.Game, source *game.Card) bool {
		return effects.AttachedToSource(target, g, source) && target.IsLegendary()
	}
	effects.Register(effects.Spec{
		OracleID:     "fbfd7f38-be9a-49bf-8769-79e1e7e8dfdc",
		Name:         "Combat Research",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []effects.AbilityGrant{{
			Key:       combatResearchDraw,
			Triggered: []game.TriggeredAbility{effects.CatgeistDraw("Combat Research — draw a card")},
			Text:      "Whenever this creature deals combat damage to a player, draw a card.",
		}},
		Static: []game.StaticAbility{
			effects.GrantAbilitiesToAttached(combatResearchDraw),
			effects.B43PumpAttachedWhen(legendary, 1, 1),
		},
		Triggered: []game.TriggeredAbility{
			effects.WardGranted(effects.WardMana("{1}"), "Combat Research — ward {1}", legendary),
		},
	})
}

const combatResearchDraw = "combat-research/draw"
