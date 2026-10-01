package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Malignus — Creature — Elemental Spirit {3}{R}{R}, */*:
//
//	"Malignus's power and toughness are each equal to half the highest
//	 life total among your opponents, rounded up.
//	 Damage that would be dealt by this creature can't be prevented."
//
// The first line is a layer 7a characteristic-defining ability (CR
// 604.3), read on every layer pass, so it moves with the life totals;
// an opponent who has left the game has no life total to count. A
// negative highest total makes it 0/0 — half of a negative number
// rounded up is never above zero. The second line is Excruciator's
// static (ADR 0107 §5).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "f90174b3-effc-4c81-8cb0-856a0646516d",
		Name:         "Malignus",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7A_CDA,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := malignusSize(g, source)
				c.Power, c.Toughness = n, n
			},
		}},
		DamageCantBePrevented: DamageByThisCantBePrevented(),
	})
}

// malignusSize is half the highest life total among `source`'s
// opponents still in the game, rounded up, and never below zero.
func malignusSize(g *game.Game, source *game.Card) int {
	best, seen := 0, false
	for _, p := range g.Seats {
		if p == nil || p.ID == source.Controller || p.Eliminated {
			continue
		}
		if !seen || p.Life > best {
			best, seen = p.Life, true
		}
	}
	if !seen || best <= 0 {
		return 0
	}
	return (best + 1) / 2
}
