package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
		return AttachedToSource(target, g, source) && target.IsLegendary()
	}
	Register(Spec{
		OracleID:     "fbfd7f38-be9a-49bf-8769-79e1e7e8dfdc",
		Name:         "Combat Research",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []AbilityGrant{{
			Key:       combatResearchDraw,
			Triggered: []game.TriggeredAbility{catgeistDraw("Combat Research — draw a card")},
			Text:      "Whenever this creature deals combat damage to a player, draw a card.",
		}},
		Static: []game.StaticAbility{
			GrantAbilitiesToAttached(combatResearchDraw),
			b43PumpAttachedWhen(legendary, 1, 1),
		},
		Triggered: []game.TriggeredAbility{
			WardGranted(WardMana("{1}"), "Combat Research — ward {1}", legendary),
		},
	})
}

const combatResearchDraw = "combat-research/draw"
