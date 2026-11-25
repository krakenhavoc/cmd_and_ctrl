package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cairn Wanderer — {4}{B} Creature — Shapeshifter 4/4 (Commander 2020):
//
//	"Changeling
//	 As long as a creature card with flying is in a graveyard, this creature
//	 has flying. The same is true for fear, first strike, double strike,
//	 deathtouch, haste, landwalk, lifelink, protection, reach, trample,
//	 shroud, and vigilance."
//
// Soulflayer's static over every graveyard instead of the delved cards. Every
// creature card in any graveyard lends the ability tokens it has: the plain
// keywords of the list, any landwalk token (the same landwalk, e.g.
// swampwalk) and any protection token (protection from the same quality), as
// the card's rulings give them. Graveyards changing re-run the layer pass
// (the graveyard arrival bump), so the Wanderer gains and loses them as cards
// come and go.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a549d54f-2640-455c-8a93-bef4df2f9e8b",
		Name:            "Cairn Wanderer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"changeling"},
		Static: []game.StaticAbility{{
			Label:     "As long as a creature card with flying (or fear, first strike, double strike, deathtouch, haste, landwalk, lifelink, protection, reach, trample, shroud, vigilance) is in a graveyard, this creature has that ability.",
			Layer:     game.Layer6Ability,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, _ *game.Card) {
				for _, p := range g.Seats {
					if p == nil || p.Graveyard == nil {
						continue
					}
					for i := range p.Graveyard.Cards {
						card := &p.Graveyard.Cards[i]
						if !card.IsCreature() {
							continue
						}
						for _, tok := range game.AbilityTokensOf(card) {
							if cairnWandererLends(tok) {
								c.Abilities = game.AppendKeywordAbility(c.Abilities, tok)
							}
						}
					}
				}
			},
		}},
	})
}

var cairnWandererKeywords = map[string]bool{
	"flying": true, "fear": true, "first strike": true, "double strike": true, "deathtouch": true,
	"haste": true, "lifelink": true, "reach": true, "trample": true, "shroud": true, "vigilance": true,
}

func cairnWandererLends(tok string) bool {
	if cairnWandererKeywords[tok] || game.IsLandwalkToken(tok) {
		return true
	}
	_, ok := game.ParseProtectionQuality(tok)
	return ok
}
