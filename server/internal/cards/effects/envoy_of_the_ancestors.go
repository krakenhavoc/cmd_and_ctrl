package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Envoy of the Ancestors — Creature — Human Cleric {2}{W}, 2/3 (Modern
// Horizons 3):
//
//	"Outlast {W} ({W}, {T}: Put a +1/+1 counter on this creature. Outlast
//	 only as a sorcery.)
//	 Modified creatures you control have lifelink. (Equipment, Auras you
//	 control, and counters are modifications.)"
//
// Outlast is the shared sorcery-speed ability (rprB04Outlast). The
// lifelink is Kodama of the West Tree's static over modified creatures
// you control (b07IsModified: a counter of any kind, an Equipment, or an
// Aura you control attached), the Envoy itself included once it carries a
// counter; it is layer 6 and re-evaluated on every pass, so lifelink comes
// and goes with the modification.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "678ef5e4-800d-4d3a-8761-2abaad70e0c9",
		Name:         "Envoy of the Ancestors",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{rprB04Outlast("{W}")},
		Static: []game.StaticAbility{{
			Layer: game.Layer6Ability,
			Label: "Modified creatures you control have lifelink.",
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller && b07IsModified(g, *target)
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				if !eotHasAbility(c.Abilities, "lifelink") {
					c.Abilities = append(c.Abilities, "lifelink")
				}
			},
		}},
	})
}
