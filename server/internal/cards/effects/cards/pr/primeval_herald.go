package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Primeval Herald — Creature — Elf Scout {3}{G}, 3/1 (Foundations
// Jumpstart):
//
//	"Trample
//	 Whenever this creature enters or attacks, you may search your library
//	 for a basic land card, put it onto the battlefield tapped, then
//	 shuffle."
//
// One printed trigger, so one registered trigger on both events (Aetherwind
// Basker's shape). The "you may" is the trigger's prompt; the search is
// Aerial Surveyor's, for any basic land.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "495348d7-bc11-4209-8b0f-72db2e6c6993",
		Name:            "Primeval Herald",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.TriggerWithPurpose(
				effects.WhenThisEntersOrAttacks("Primeval Herald — you may search for a basic land card, put it onto the battlefield tapped",
					func(g *game.Game, item *game.StackItem) error {
						return effects.SearchLibrary{
							Player:        item.Controller,
							Predicate:     effects.IsBasicLand,
							Dest:          game.ZoneBattlefield,
							Limit:         1,
							Reveal:        false,
							Shuffle:       true,
							TappedOnEntry: true,
							Reason:        "Primeval Herald — a basic land card, onto the battlefield tapped",
						}.Apply(effects.NewContext(g, item))
					}), game.Purpose{Lands: 1}),
				"Primeval Herald — search for a basic land card?"),
		},
	})
}
