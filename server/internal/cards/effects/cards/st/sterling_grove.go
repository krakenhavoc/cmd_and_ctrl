package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sterling Grove — Enchantment {G}{W} (Modern Horizons 2, #5501):
//
//	"Other enchantments you control have shroud.
//	 {1}, Sacrifice this enchantment: Search your library for an
//	 enchantment card, reveal it, then shuffle and put that card on top."
//
// The shroud grant is a layer 6 keyword grant over the controller's other
// enchantments (CR 702.18a); Sterling Grove itself is not one of them. The
// activated ability is the one-mana tutors' search: reveal, shuffle, then
// the card goes on top, and a search that finds nothing still shuffles.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2c275a85-5a15-46cd-a6e7-add63f9b853d",
		Name:         "Sterling Grove",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID != source.InstanceID && target.IsEnchantment() && target.Controller == source.Controller
			}, "shroud"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, Sacrifice this enchantment: Search your library for an enchantment card, reveal it, then shuffle and put that card on top.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Tutors: 1},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.TutorToTop(effects.NewContext(g, item), "Sterling Grove — an enchantment card", true,
					func(c game.Card) bool { return c.IsEnchantment() })
			},
		}},
	})
}
