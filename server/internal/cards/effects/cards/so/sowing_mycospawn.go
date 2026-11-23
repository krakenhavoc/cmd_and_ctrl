package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sowing Mycospawn — Creature — Eldrazi Fungus {3}{G}, 3/3 (Modern Horizons 3):
//
//	"Devoid
//	 Kicker {1}{C}
//	 When you cast this spell, search your library for a land card, put it onto
//	 the battlefield, then shuffle.
//	 When you cast this spell, if it was kicked, exile target land."
//
// Two cast triggers: the search (untapped, any land card) and the kicked
// exile, both on the stack above the spell.
func init() {
	const search = "Sowing Mycospawn — search your library for a land card, put it onto the battlefield, then shuffle"
	effects.Register(effects.Spec{
		OracleID:        "63a19c7c-49d7-4b75-aac8-d0ff36c45fb8",
		Name:            "Sowing Mycospawn",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		OptionalCosts:   []game.AdditionalCost{effects.Kicker("{1}{C}")},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.WhenYouCastThisSpell(search, func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player:    item.Controller,
					Predicate: func(c game.Card) bool { return c.IsLand() },
					Dest:      game.ZoneBattlefield,
					Limit:     1,
					Shuffle:   true,
					Reason:    "Sowing Mycospawn — a land card, onto the battlefield",
				}.Apply(effects.NewContext(g, item))
			}), game.Purpose{Lands: 1}),
			effects.CastTriggerIfKickedWith("{1}{C}", "Sowing Mycospawn — kicked, exile target land",
				effects.TargetPermanent("target land", effects.Land()), effects.ExileTheFirstLegalTarget),
		},
	})
}
