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
// Fully implemented as printed. The Layer 6 static below grants
// "hexproof" to every other tapped creature you control, and since
// S23 the targeting choke point (game.CanBeTargetedBy, called from
// targets.go at both the CR 601.2c announce gate and the CR 608.2b
// resolution re-check) reads it: an opponent's removal spell cannot
// be announced at one of those creatures, and one already on the
// stack fizzles if its target becomes tapped — and therefore
// hexproof — in response.
//
// Shipped S22 with the grant declared but INERT, because hexproof
// was not yet in the engine's enforced keyword table; the note that
// said so was removed when the table gained it. Nothing about this
// card had to change for it to start working, which was the point
// of declaring the grant rather than omitting it.
//
// The grant is a Layer 6 ability-adding effect keyed on a state the
// layer engine recomputes (Tapped), so it comes and goes with the
// tap: untapping the creature removes the hexproof on the next
// recompute, exactly as printed.
//
// Reviewed line by line for #1306: flash and double strike are
// printed keywords, convoke is Spec.TapCost, and the hexproof grant
// covers OTHER, TAPPED CREATURES YOU CONTROL — each word a clause of
// AppliesTo. Hexproof stops only opponents (CR 702.11b), which the
// targeting gate enforces. No simplifications.
func init() {
	Register(Spec{
		OracleID:        "b8ef65df-f8e7-44e3-9864-9c127232a2b6",
		Name:            "The Wandering Rescuer",
		Completeness:    CompletenessFull,
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
