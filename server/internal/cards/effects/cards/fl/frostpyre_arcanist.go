package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Frostpyre Arcanist — Creature — Giant Wizard {4}{U}, 2/5 (Kaldheim):
//
//	"This spell costs {1} less to cast if you control a Giant or a Wizard.
//	 When this creature enters, search your library for an instant or sorcery
//	 card with the same name as a card in your graveyard, reveal it, put it
//	 into your hand, then shuffle."
//
// The names are read from the controller's graveyard as the trigger resolves
// (any card there, not only instants and sorceries); the search names a
// quality, so it may find nothing (CR 701.23b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6bdb251b-51fe-46cc-8b40-540d081c9419",
		Name:         "Frostpyre Arcanist",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLess(1, "This spell costs {1} less to cast if you control a Giant or a Wizard.",
				effects.YouControlA(effects.Or(effects.HasSubtype("Giant"), effects.HasSubtype("Wizard")))),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Frostpyre Arcanist — search your library for an instant or sorcery card with the same name as a card in your graveyard, reveal it, put it into your hand, then shuffle",
				func(g *game.Game, item *game.StackItem) error {
					names := map[string]bool{}
					if p := g.PlayerByIDForEffect(item.Controller); p != nil {
						for _, c := range p.Graveyard.Cards {
							names[c.Name] = true
						}
					}
					return effects.SearchLibrary{
						Player: item.Controller,
						Predicate: func(c game.Card) bool {
							return (c.IsInstant() || c.IsSorcery()) && names[c.Name]
						},
						Dest:    game.ZoneHand,
						Limit:   1,
						Reveal:  true,
						Shuffle: true,
						Reason:  "Frostpyre Arcanist — an instant or sorcery card with the same name as a card in your graveyard",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
