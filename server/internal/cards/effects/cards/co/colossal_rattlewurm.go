package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Colossal Rattlewurm — Creature — Wurm {2}{G}{G}, 6/5 (Outlaws of Thunder Junction):
//
//	"Colossal Rattlewurm has flash as long as you control a Desert.
//	 Trample
//	 {1}{G}, Exile this card from your graveyard: Search your library for a
//	 Desert card, put it onto the battlefield tapped, then shuffle."
//
// The flash is a timing permission read as the cast begins (Spec.FlashIf,
// CR 601.3): a Desert you control, any card with the subtype. The graveyard
// ability is activated from the graveyard (Zones, CR 113.6), exiles the card as
// a cost, and fetches like Wight of the Reliquary with a Desert in place of
// any land.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "510659d6-69b4-4b04-8836-08b76213954e",
		Name:            "Colossal Rattlewurm",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Purpose:         game.Purpose{Lands: 1},
		FlashIf: effects.FlashIfYou("as long as you control a Desert", func(g *game.Game, you uuid.UUID) bool {
			for _, c := range g.BattlefieldCardsForEffect() {
				if c.Controller == you && c.HasSubtype("Desert") {
					return true
				}
			}
			return false
		}),
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{G}, Exile this card from your graveyard: Search your library for a Desert card, put it onto the battlefield tapped, then shuffle.",
			Cost:    effects.Plus(effects.ManaCost("{1}{G}"), effects.ExileThis()),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Purpose: game.Purpose{Lands: 1, Answers: game.AnswerValue},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player:        item.Controller,
					Predicate:     func(c game.Card) bool { return c.HasSubtype("Desert") },
					Dest:          game.ZoneBattlefield,
					Limit:         1,
					Reveal:        true,
					Shuffle:       true,
					TappedOnEntry: true,
					Reason:        "Colossal Rattlewurm — a Desert card, onto the battlefield tapped",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
