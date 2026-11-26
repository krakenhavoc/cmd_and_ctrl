package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// From Beyond — Enchantment {3}{G} (Battle for Zendikar, #5547):
//
//	"Devoid (This card has no color.)
//	 At the beginning of your upkeep, create a 1/1 colorless Eldrazi Scion
//	 creature token. It has "Sacrifice this token: Add {C}."
//	 {1}{G}, Sacrifice this enchantment: Search your library for an Eldrazi
//	 card, reveal it, put it into your hand, then shuffle."
//
// The upkeep trigger is Scion Summoner's token. The search is Eye of Ugin's,
// for any Eldrazi card (a type word on a card in the library, so a changeling
// is one too).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "31b27775-196b-4022-bff5-25b34290ff7f",
		Name:            "From Beyond",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.AtYourUpkeep("From Beyond — create a 1/1 colorless Eldrazi Scion",
				effects.Do(effects.CreateToken{Template: effects.B28EldraziScionToken(), N: 1})),
				game.Purpose{Tokens: 1}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{G}, Sacrifice this enchantment: Search your library for an Eldrazi card, reveal it, put it into your hand, then shuffle.",
			Purpose: game.Purpose{Tutors: 1, Answers: game.AnswerValue},
			Cost:    game.AbilityCost{SacrificeSelf: true, Mana: "{1}{G}"},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				eldrazi := ctx.TextSubtype("Eldrazi")
				return effects.SearchLibrary{
					Player:    item.Controller,
					Predicate: func(c game.Card) bool { return c.HasSubtype(eldrazi) },
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "From Beyond — an Eldrazi card",
				}.Apply(ctx)
			},
		}},
	})
}
