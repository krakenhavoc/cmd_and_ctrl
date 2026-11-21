package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sigarda, Heron's Grace — Legendary Creature — Angel {3}{G}{W}, 4/5:
//
//	"Flying
//	 You and Humans you control have hexproof.
//	 {2}, Exile a card from your graveyard: Create a 1/1 white Human
//	 Soldier creature token."
//
// "You have hexproof" is a player keyword and the Humans are a Layer 6
// grant over the Humans you control (Sigarda is an Angel, so she is
// not one unless something makes her so). The exile is a cost, chosen
// by the activator at announce.
//
// No simplification.
func init() {
	grant := effects.KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
		return effects.TribeFilter{Tribes: []string{"Human"}, YoursOnly: true}.Matches(target, g, source)
	}, effects.KeywordHexproof)
	grant.Label = "Humans you control have hexproof."

	effects.Register(effects.Spec{
		OracleID:        "79b56cf8-d573-40cf-b797-6b4eef2bed23",
		Name:            "Sigarda, Heron's Grace",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		PlayerKeywords:  []string{effects.KeywordHexproof},
		Static:          []game.StaticAbility{grant},
		Activated: []effects.ActivatedAbility{{
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Label:   "{2}, Exile a card from your graveyard: Create a 1/1 white Human Soldier creature token.",
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.ExileFromGraveyard(1, "a card", nil)),
			Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Human Soldier"), N: 1}),
		}},
	})
}
