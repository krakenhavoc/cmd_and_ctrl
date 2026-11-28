package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Alexi, Zephyr Mage — Legendary Creature — Human Spellshaper {3}{U}{U}, 3/3
// (Prophecy):
//
//	"{X}{U}, {T}, Discard two cards: Return X target creatures to their
//	 owners' hands."
//
// Icy Blast's X targets (exactly X, chosen as the ability is activated; the count is the target clause's, so no XMatters) behind
// Greel's cost shape; each target still legal as it resolves goes back to its
// owner's hand (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3f60de36-ed63-4d08-a012-fc16e91da46d",
		Name:         "Alexi, Zephyr Mage",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{X}{U}, {T}, Discard two cards: Return X target creatures to their owners' hands.",
			Cost:    effects.Plus(effects.ManaCost("{X}{U}"), effects.TapCost(), effects.DiscardN(2, "two cards")),
			Targets: effects.OldB347XTargetCreatures(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.NeoB20BounceTargetedPermanents(effects.NewContext(g, item))
			},
		}},
	})
}
