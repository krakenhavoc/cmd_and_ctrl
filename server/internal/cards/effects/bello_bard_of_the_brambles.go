package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bello, Bard of the Brambles — Legendary Creature — Raccoon Bard {1}{R}{G},
// 3/3 (Bloomburrow Commander, #5368):
//
//	"During your turn, each non-Equipment artifact and non-Aura enchantment
//	 you control with mana value 4 or greater is a 4/4 Elemental creature in
//	 addition to its other types and has indestructible, haste, and
//	 'Whenever this creature deals combat damage to a player, draw a card.'"
//
// Four layered statics (cmdr2S18BelloStatics): the type in layer 4, the
// keywords and the keyed trigger grant in layer 6, and base 4/4 in layer 7b.
// All of them read "during your turn" and the mana value as they apply, so
// the permanents are plain noncreature permanents on every other turn.
func init() {
	const grantKey = "cmdr2-s18/bello-combat-damage-draw"
	Register(Spec{
		OracleID:     "fb773114-6161-46a3-ab30-1853958fbcb0",
		Name:         "Bello, Bard of the Brambles",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: grantKey,
			Triggered: []game.TriggeredAbility{
				WheneverThisDealsCombatDamageToAPlayer("Bello — draw a card", Do(DrawCards{N: 1})),
			},
			Text: "Whenever this creature deals combat damage to a player, draw a card.",
		}},
		Static: cmdr2S18BelloStatics(grantKey),
	})
}
