package cardslu

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Meddling Mage — Creature — Human Wizard {W}{U}, 2/2 (Alara Reborn, #5586):
//
//	"As this creature enters, choose a nonland card name.
//	 Spells with the chosen name can't be cast."
//
// Nevermore's entry choice and cast bar on a creature: the name is chosen as it
// enters (CR 614.12) through the shared card-name prompt, per instance, and
// matched on every face (CR 201.2b). A land can never be a spell, so a land's
// name bars nothing; the prompt cannot check a typed name against a real card
// (Pithing Needle's caveat), which a paper table does by announcing it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a5d02339-9f6d-486e-87d8-5ec2274cf680",
		Name:         "Meddling Mage",
		Completeness: effects.CompletenessFull,
		AsEnters:     effects.ChooseCardNameAsEnters("Meddling Mage — choose a nonland card name"),
		CastRestrictions: []game.CastRestriction{{
			Label: "Spells with the chosen name can't be cast.",
			Forbids: func(q game.CastQuery) bool {
				named := strings.TrimSpace(q.Source.ChosenName)
				return named != "" && game.CardNameMatches(q.Card, named)
			},
		}},
	})
}
