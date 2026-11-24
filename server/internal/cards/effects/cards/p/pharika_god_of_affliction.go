package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pharika, God of Affliction — Legendary Enchantment Creature — God {1}{B}{G},
// 5/5 (Secret Lair Drop):
//
//	"Indestructible
//	 As long as your devotion to black and green is less than seven, Pharika
//	 isn't a creature.
//	 {B}{G}: Exile target creature card from a graveyard. Its owner creates a
//	 1/1 black and green Snake enchantment creature token with deathtouch."
//
// The God clause is the shared two-colour one (GodUnlessDevotionToColors:
// black and green, seven). The token is made by the exiled card's OWNER, who
// is whoever's graveyard it was in, even an opponent's; a card that left the
// graveyard in response makes no token (the ability's target is illegal).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b7d05d00-1e69-4d11-bff0-c1ca2026aad6",
		Name:            "Pharika, God of Affliction",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		Static:          []game.StaticAbility{effects.GodUnlessDevotionToColors(7, "B", "G")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}{G}: Exile target creature card from a graveyard. Its owner creates a 1/1 black and green Snake enchantment creature token with deathtouch.",
			Cost:    effects.ManaCost("{B}{G}"),
			Targets: effects.TargetCardInGraveyard("target creature card from a graveyard", effects.Creature()),
			Purpose: game.Purpose{Answers: game.AnswerGraveyard, Tokens: 1, CreatureTokens: 1},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, id := range effects.LegalTargetCardIDs(ctx) {
					c, ok := g.LookupCardForEffect(id)
					if !ok {
						continue
					}
					owner := c.Owner
					if err := (effects.ExileTarget{Target: id}).Apply(ctx); err != nil {
						return err
					}
					return effects.CreateToken{Controller: owner, Template: effects.TokenCard("1/1 black and green Snake enchantment with deathtouch"), N: 1}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
