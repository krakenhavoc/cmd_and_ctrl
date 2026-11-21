package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jenson Carthalion, Druid Exile — Legendary Creature — Human Druid {G}{W},
// 2/2 (Dominaria United, S82 DMU batch 15):
//
//	"Whenever you cast a multicolored spell, scry 1. If that spell was all
//	 colors, create a 4/4 white Angel creature token with flying and
//	 vigilance.
//	 {5}, {T}: Add {W}{U}{B}{R}{G}."
//
// One trigger: it scries 1 for every multicolored spell, then, if that spell
// had all five colors (read from the spell as it was cast), creates the
// Angel. The last line is a mana ability with a mana cost (the Signets'
// shape): {5} and {T} for one of each color.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "68cefb44-afad-4638-8fdc-65a099b51d55",
		Name:         "Jenson Carthalion, Druid Exile",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Scry: 1},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Multicolored(),
				"Jenson Carthalion — scry 1; if that spell was all colors, create a 4/4 white Angel creature token with flying and vigilance",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if err := (effects.Scry{N: 1}).Apply(ctx); err != nil {
						return err
					}
					spell, ok := g.LookupCardForEffect(ctx.Trigger().Event.CardID)
					if !ok || !effects.IsAllColors(spell) {
						return nil
					}
					return effects.CreateToken{Template: effects.TokenCard("4/4 white Angel with flying and vigilance"), N: 1}.Apply(ctx)
				}),
		},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true, Mana: "{5}"},
			Produced: "{W}{U}{B}{R}{G}",
			Label:    "{5}, {T}: Add {W}{U}{B}{R}{G}.",
		}},
	})
}
