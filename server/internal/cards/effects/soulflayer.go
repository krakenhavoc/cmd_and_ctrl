package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Soulflayer — Creature — Demon 4/4 {4}{B}{B}:
//
//	"Delve
//	 If a creature card with flying was exiled with this creature's
//	 delve ability, this creature has flying. The same is true for
//	 first strike, double strike, deathtouch, haste, hexproof,
//	 indestructible, lifelink, reach, trample, and vigilance."
//
// ADR 0100 sub-PR 2. "Exiled with this creature's delve ability" is CR
// 607.2q's link, carried onto the permanent as game.Card.Delved and
// resolved through Game.DelvedCardsForEffect: the cards delve exiled to
// pay for the spell that became this Soulflayer, and only while each is
// still in exile as that object. A card that leaves exile is a new
// object (CR 400.7) and stops counting, so the static declares
// DependsOnExile and the layer engine re-reads it when one leaves.
//
// It is a layer-6 self-grant. Each linked card is asked as it sits in
// exile, where it has only its printed characteristics: the 2014-11-24
// ruling is that a creature card which grants itself a keyword ("as
// long as you control a red or white permanent, Battle Brawler … has
// first strike") gives Soulflayer nothing, which is what reading the
// card's own keyword list does.
//
// Simplification: the 2020-01-24 ruling gives Soulflayer a keyword's
// VARIANT when the exiled card has one — "hexproof from white". The
// engine has no hexproof-from keyword (ADR 0038 §6), so such a card
// gives Soulflayer nothing. That is weaker than printed, never
// stronger.
func init() {
	Register(Spec{
		OracleID:     "d45cf9b1-d916-48bf-8b99-60cd09230a4b",
		Name:         "Soulflayer",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"A creature card with hexproof from a color or quality doesn't give Soulflayer that ability.",
		},
		Delve:  true,
		Static: []game.StaticAbility{soulflayerKeywordsFromDelvedCards()},
	})
}

// soulflayerKeywords are the eleven keywords Soulflayer's static
// names, as the engine's canonical tokens.
var soulflayerKeywords = []string{
	"flying", "first strike", "double strike", "deathtouch", "haste", "hexproof",
	"indestructible", "lifelink", "reach", "trample", "vigilance",
}

// soulflayerKeywordsFromDelvedCards is the static: for each keyword on
// the list, this creature has it if a creature card among the cards
// still linked to its delve has it.
func soulflayerKeywordsFromDelvedCards() game.StaticAbility {
	return game.StaticAbility{
		Layer:          game.Layer6Ability,
		AppliesTo:      selfOnly,
		DependsOnExile: true,
		Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
			linked := g.DelvedCardsForEffect(source.Delved())
			for _, kw := range soulflayerKeywords {
				for i := range linked {
					if linked[i].IsCreature() && game.HasKeyword(&linked[i], kw) {
						c.Abilities = game.AppendKeywordAbility(c.Abilities, kw)
						break
					}
				}
			}
		},
	}
}
