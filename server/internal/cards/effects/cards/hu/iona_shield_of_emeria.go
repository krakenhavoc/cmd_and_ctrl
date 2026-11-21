package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Iona, Shield of Emeria — Legendary Creature — Angel {6}{W}{W}{W}, 7/7:
//
//	"Flying
//	 As Iona enters, choose a color.
//	 Your opponents can't cast spells of the chosen color."
//
// The stored colour pick (ChooseColorAsEnters, answer on Card.ChosenColor)
// read back by an OpponentsCantCast-shaped restriction. Until the pick is
// made the colour is empty and nothing is forbidden (an empty colour never
// reads as "any colour"). The restriction is Iona's controller's: a
// stolen Iona restricts the new controller's opponents.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5e816472-561d-43d1-ad28-80a829e2a91a",
		Name:            "Iona, Shield of Emeria",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		AsEnters:        effects.ChooseColorAsEnters(game.ColorForHarm, "Iona, Shield of Emeria"),
		CastRestrictions: []game.CastRestriction{{
			Label: "Your opponents can't cast spells of the chosen color.",
			Forbids: func(q game.CastQuery) bool {
				if q.Source.Controller == q.Controller || q.Source.ChosenColor == "" {
					return false
				}
				return q.Card.HasColor(q.Source.ChosenColor)
			},
		}},
	})
}
