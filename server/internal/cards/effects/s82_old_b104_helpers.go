package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// DefenderWaiverStatic is the layer-6 static every "can attack as though it
// didn't have defender" card writes (ADR 0106's 2026-10-16 amendment, #1936):
// an unrestricted DefenderWaiver on each permanent `applies` admits. The
// keyword stays (CR 702.3b is waived, not removed). Platypus-Bear and
// Starfleet Spanner Jockey admit only themselves while a condition holds
// (AvB18CanAttackAsThoughNoDefenderWhile); Warmonger's Chariot admits the
// equipped creature.
func DefenderWaiverStatic(label string, applies func(target *game.Card, g *game.Game, source *game.Card) bool) game.StaticAbility {
	return game.StaticAbility{
		Label:     label,
		Layer:     game.Layer6Ability,
		AppliesTo: applies,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
			w := game.DefenderWaiver{}
			if source != nil {
				w.Source, w.SourceName = source.InstanceID, source.Name
			}
			c.DefenderWaivers = append(c.DefenderWaivers, w)
		},
	}
}

// EquippedCanAttackAsThoughNoDefender is "As long as equipped creature has
// defender, it can attack as though it didn't have defender" (Warmonger's
// Chariot). The waiver changes nothing for a creature without defender, so it
// is written for the equipped creature unconditionally.
func EquippedCanAttackAsThoughNoDefender(label string) game.StaticAbility {
	return DefenderWaiverStatic(label, func(target *game.Card, g *game.Game, source *game.Card) bool {
		return target.IsCreature() && AttachedToSource(target, g, source)
	})
}
