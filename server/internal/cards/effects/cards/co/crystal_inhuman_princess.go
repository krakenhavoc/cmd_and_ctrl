package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crystal, Inhuman Princess — Legendary Creature — Inhuman Noble Hero
// {1}{R}{G}, 2/3 (Marvel Super Heroes Commander):
//
//	"Flying
//	 Whenever you cast a noncreature spell, Crystal deals X damage to each
//	 opponent, where X is the number of colors that spell is.
//	 {T}: Add {R}, {G}, {W}, or {U}."
//
// X is the cast spell's colours as it was cast (a colourless spell is
// zero and deals nothing). The mana ability is a pick among the four
// named colours, not the usual five.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "32c14d7f-fdaa-4340-af94-1f10bf064f51",
		Name:            "Crystal, Inhuman Princess",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{R|G|W|U}",
			Label:    "Add {R}, {G}, {W}, or {U}",
		}},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.B10NoncreatureSpellCastByYou,
				"Crystal, Inhuman Princess — damage to each opponent equal to the spell's colors",
				func(g *game.Game, item *game.StackItem) error {
					spell, ok := g.LookupCardForEffect(item.Trigger.Event.CardID)
					if !ok {
						return nil
					}
					n := effects.ColorsOf(spell)
					if n <= 0 {
						return nil
					}
					return effects.DamageToEachOpponent(g, item, n)
				}),
		},
	})
}
