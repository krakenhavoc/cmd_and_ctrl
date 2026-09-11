package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Wandering Rescuer — Legendary Creature — Human Samurai Noble
// {3}{W}{W}, 3/4:
//
//	"Flash
//	 Convoke (Your creatures can help cast this spell. Each creature
//	 you tap while casting this spell pays for {1} or one mana of
//	 that creature's color.)
//	 Double strike
//	 Other tapped creatures you control have hexproof."
//
// Flash plus convoke is the whole card: a five-drop that a wide
// white board casts for one real mana at the end of an opponent's
// turn, and the creatures it taps to do it are then protected by the
// body that arrived. In paper the convoke and the hexproof clause
// are one loop — you tap the team, they become hexproof, and the
// sweeper aimed at them misses.
//
// S22 sandbox simplification — **the hexproof grant is inert.**
// "Hexproof" is not one of the twelve keywords the engine's
// targeting and combat code honours (PrintedKeywords accepts flying,
// reach, deathtouch, lifelink, trample, vigilance, first strike,
// double strike, menace, defender, haste, flash), so the static
// ability below really does append the string to every other tapped
// creature you control, and nothing reads it. The grant is declared
// rather than omitted so that the day hexproof lands in the
// targeting gate this card starts working without being touched.
//
// The simplification is strictly WEAKER than printed: the creatures
// convoked to cast this are targetable when paper says they would
// not be. The flash, the convoke, and the double strike are all
// real.
func init() {
	Register(Spec{
		OracleID:        "b8ef65df-f8e7-44e3-9864-9c127232a2b6",
		Name:            "The Wandering Rescuer",
		PrintedKeywords: []string{"flash", "double strike"},
		TapCost:         Convoke(),
		Static: []game.StaticAbility{{
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.InstanceID != source.InstanceID &&
					target.Controller == source.Controller &&
					target.Tapped && target.IsCreature()
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				for _, a := range c.Abilities {
					if a == "hexproof" {
						return
					}
				}
				c.Abilities = append(c.Abilities, "hexproof")
			},
		}},
	})
}
