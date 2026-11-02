package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kraven, Proud Predator — Legendary Creature — Human Warrior Villain
// {1}{R}{G}, */4 (Marvel's Spider-Man, #4460):
//
//	"Vigilance
//	 Top of the Food Chain — Kraven's power is equal to the greatest mana
//	 value among permanents you control."
//
// A layer 7a characteristic-defining ability (Lord of Extinction's shape)
// that sets power only; the printed 4 toughness stays. Kraven's own mana
// value (3) is among the permanents he counts, and a token with no mana
// cost counts 0.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "12b8b4bf-ba85-419c-8005-1ec34b2b1deb",
		Name:            "Kraven, Proud Predator",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Static: []game.StaticAbility{{
			Label:    "Top of the Food Chain — Kraven's power is equal to the greatest mana value among permanents you control.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7A_CDA,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				best := 0
				for _, p := range g.BattlefieldCardsForEffect() {
					if p.Controller != source.Controller {
						continue
					}
					if mv := p.ManaValue(); mv > best {
						best = mv
					}
				}
				c.Power = best
			},
		}},
	})
}
