package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vishgraz, the Doomhive — Legendary Creature — Phyrexian Insect {2}{W}{B}{G},
// 3/3:
//
//	"Menace, toxic 1
//	 When Vishgraz enters, create three 1/1 colorless Phyrexian Mite artifact
//	 creature tokens with toxic 1 and 'This token can't block.'
//	 Vishgraz gets +1/+1 for each poison counter your opponents have."
//
// The bonus is a layer 7c static summing every opponent's poison counters on
// each recompute, so it moves as counters come and go.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7b1599c0-bb4f-48f8-a358-dc60166b13e7",
		Name:            "Vishgraz, the Doomhive",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace", "toxic 1"},
		Purpose:         game.Purpose{Tokens: 3, CreatureTokens: 3},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Vishgraz — create three Phyrexian Mites", effects.Do(effects.CreateToken{Template: effects.PhyrexianMiteToken(), N: 3})),
		},
		Static: []game.StaticAbility{{
			Label:     "Vishgraz gets +1/+1 for each poison counter your opponents have.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: effects.SelfCreature,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := effects.S82OneB04OpponentPoison(g, source.Controller)
				c.Power += n
				c.Toughness += n
			},
		}},
	})
}
