package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shardmage's Rescue — Enchantment — Aura {W} (Duskmourn: House of
// Horror, #4466):
//
//	"Flash
//	 Enchant creature you control
//	 As long as this Aura entered this turn, enchanted creature has
//	 hexproof.
//	 Enchanted creature gets +1/+1."
//
// Aquitect's Defenses' flash Aura on a creature you control, with the
// hexproof as a static rather than an enters trigger: a layer 6 grant on
// the enchanted creature (GrantToAttachedWhile) while the Aura entered
// this turn, Drownyard Behemoth's condition (Game.EnteredThisTurn, not
// summoning sickness). Cast in an opponent's turn, the creature has
// hexproof for the rest of that turn and loses it as the next turn
// begins; the turn boundary invalidates the layer cache. Moving the Aura
// to another creature keeps the grant with the Aura that turn, and the
// +1/+1 is PumpAttached.
//
// No simplification.
func init() {
	hexproof := GrantToAttachedWhile(func(_ *game.Card, g *game.Game, source *game.Card) bool {
		return g.EnteredThisTurn(source.InstanceID)
	}, "hexproof")
	hexproof.Label = "As long as this Aura entered this turn, enchanted creature has hexproof."
	Register(Spec{
		OracleID:        "406247ee-2b86-4068-b5fa-65cdc83a82f4",
		Name:            "Shardmage's Rescue",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Purpose:         game.Purpose{Aura: game.AuraBenefit},
		Targets:         EnchantCreature(YouControl()),
		Static: []game.StaticAbility{
			hexproof,
			PumpAttached(1, 1),
		},
	})
}
