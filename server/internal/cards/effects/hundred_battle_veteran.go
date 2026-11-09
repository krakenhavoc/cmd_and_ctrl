package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hundred-Battle Veteran — Creature — Zombie Warrior {3}{B}, 4/2 (Tarkir:
// Dragonstorm, #4463):
//
//	"As long as there are three or more different kinds of counters among
//	 creatures you control, this creature gets +2/+4.
//	 You may cast this card from your graveyard. If you do, it enters with
//	 a finality counter on it. (If a creature with a finality counter on
//	 it would die, exile it instead.)"
//
// The boost is a layer 7c static that re-reads the board on every recompute:
// it counts the distinct counter kinds (any kind, a finality or stun counter
// included) over the creatures its controller controls, itself too. The
// graveyard cast and its finality counter are Storied Veteran's
// (CastableZones, CountersIfCastFrom).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "9ed6aa5b-4369-42f7-9853-aae1ac9b17e2",
		Name:          "Hundred-Battle Veteran",
		Completeness:  CompletenessFull,
		CastableZones: []game.ZoneKind{game.ZoneGraveyard},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{
			CountersIfCastFrom(game.ZoneGraveyard, game.CounterFinality, 1),
		},
		Static: []game.StaticAbility{{
			Label:    "As long as there are three or more different kinds of counters among creatures you control, this creature gets +2/+4.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && countersKindsAmongYourCreatures(g, source.Controller) >= 3
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power += 2
				c.Toughness += 4
			},
		}},
	})
}

// countersKindsAmongYourCreatures is the number of different kinds of
// counters on the creatures `player` controls, as the battlefield stands.
func countersKindsAmongYourCreatures(g *game.Game, player uuid.UUID) int {
	kinds := map[string]bool{}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != player || !c.IsCreature() {
			continue
		}
		for kind, n := range c.Counters {
			if n > 0 {
				kinds[kind] = true
			}
		}
	}
	return len(kinds)
}
