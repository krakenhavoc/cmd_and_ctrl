package cardsn

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nevermore — Enchantment {1}{W}{W} (Innistrad, S82 OLD batch 450, #5572):
//
//	"As this enchantment enters, choose a nonland card name.
//	 Spells with the chosen name can't be cast."
//
// Gideon's Intervention's entry choice and cast bar, for every player
// rather than only opponents. The name is matched on every face
// (CR 201.2b), so naming either half of a split or double-faced card bars
// the card.
//
// A land can never be a spell, so a land's name bars nothing; the prompt
// cannot check a typed name against a real card (Pithing Needle's
// caveat), which a paper table does by announcing it.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0a1f4892-c9fa-4c60-a41e-7986cdd59096",
		Name:         "Nevermore",
		Completeness: effects.CompletenessFull,
		AsEnters:     effects.ChooseCardNameAsEnters("Nevermore — choose a nonland card name"),
		CastRestrictions: []game.CastRestriction{{
			Label: "Spells with the chosen name can't be cast.",
			Forbids: func(q game.CastQuery) bool {
				named := strings.TrimSpace(q.Source.ChosenName)
				return named != "" && game.CardNameMatches(q.Card, named)
			},
		}},
	})
}
