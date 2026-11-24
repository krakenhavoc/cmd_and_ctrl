package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gravebreaker Lamia — Enchantment Creature — Snake Lamia {4}{B}, 4/4 (Secret
// Lair Drop):
//
//	"Lifelink
//	 When this creature enters, search your library for a card, put it into
//	 your graveyard, then shuffle.
//	 Spells you cast from your graveyard cost {1} less to cast."
//
// The search is an any-card search to the graveyard (SearchLibrary, Dest
// graveyard, shuffled). The discount is a cost modifier for spells cast from
// the controller's graveyard (flashback, escape, a granted permission): the
// generic part of the cost drops by one and never below zero (CR 601.2f).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2c1d288a-22ea-482c-947e-f3ab04172da0",
		Name:            "Gravebreaker Lamia",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Purpose:         game.Purpose{Answers: game.AnswerGraveyard},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Gravebreaker Lamia — search your library for a card, put it into your graveyard, then shuffle",
				func(g *game.Game, item *game.StackItem) error {
					return effects.SearchLibrary{
						Player:  item.Controller,
						Dest:    game.ZoneGraveyard,
						Limit:   1,
						Shuffle: true,
						Reason:  "Gravebreaker Lamia — search your library for a card to put into your graveyard",
					}.Apply(effects.NewContext(g, item))
				}),
		},
		CostModifiers: []game.CostModifier{
			effects.CostsLess(1, "Spells you cast from your graveyard cost {1} less to cast.", effects.YourSpell(), effects.CastFromGraveyard()),
		},
	})
}
