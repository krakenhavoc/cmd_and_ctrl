package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kethis, the Hidden Hand — Legendary Creature — Elf Advisor {W}{B}{G}, 3/4
// (Core Set 2020):
//
//	"Legendary spells you cast cost {1} less to cast.
//	 Exile two legendary cards from your graveyard: Until end of turn, each
//	 legendary card in your graveyard gains "You may play this card from your
//	 graveyard.""
//
// The discount is Bard Class's legendary-spell reduction at one generic
// (CR 601.2f). The cost exiles two legendary cards from the graveyard (CR
// 602.2b). On resolution the set of legendary cards in the graveyard is fixed
// then (CR 611.2c, the cards the cost exiled are gone) and each is given a
// permission to be played from there until end of turn: a play, so a legendary
// land too, at the card's own timing and for its own cost (CR 601.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1bf08ba0-4235-4dea-9722-ed991a3b6044",
		Name:         "Kethis, the Hidden Hand",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsLess(1, "Legendary spells you cast cost {1} less to cast.", effects.YourSpell(), effects.LegendarySpell()),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Exile two legendary cards from your graveyard: Until end of turn, each legendary card in your graveyard gains \"You may play this card from your graveyard.\"",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ExileFromGraveyard(2, "two legendary cards", func(c game.Card) bool { return c.IsLegendary() }),
			Effect:  kethisGrantPlay,
		}},
	})
}

func kethisGrantPlay(g *game.Game, item *game.StackItem) error {
	p := g.PlayerByIDForEffect(item.Controller)
	if p == nil || p.Graveyard == nil {
		return nil
	}
	var cards []game.Card
	for _, c := range p.Graveyard.Cards {
		if c.IsLegendary() {
			cards = append(cards, c)
		}
	}
	g.GrantCastPermissionToCardsForEffect(game.CastPermission{
		Player:     item.Controller,
		Zone:       game.ZoneGraveyard,
		Source:     item.SourceCardID,
		SourceName: "Kethis, the Hidden Hand",
		Label:      "Play from your graveyard (Kethis, the Hidden Hand)",
	}, cards)
	return nil
}
