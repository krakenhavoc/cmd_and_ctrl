package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Leonin Lightbringer — Creature — Cat Rebel {2}{W}, 3/2 (Phyrexia: All
// Will Be One):
//
//	"Ward {2}
//	 As long as this creature is equipped, it gets +1/+1."
//
// #5479, S82. Armed Assailant's "as long as this creature is equipped" gate
// (isEquippedCreature, read on every layer recompute) on a layer 7c pump of
// the creature alone, next to the ward trigger.
//
// No simplification.
func init() {
	whileEquipped := func(target *game.Card, g *game.Game, source *game.Card) bool {
		return selfOnly(target, g, source) && isEquippedCreature(g, *target)
	}
	pump := b16Anthem(whileEquipped, 1, 1)
	pump.Label = "As long as this creature is equipped, it gets +1/+1."
	Register(Spec{
		OracleID:     "1f874f7f-b959-4de5-a5db-e65a90ac2f9d",
		Name:         "Leonin Lightbringer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{2}"), "Leonin Lightbringer — ward {2}"),
		},
		Static: []game.StaticAbility{pump},
	})
}
