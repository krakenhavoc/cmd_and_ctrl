package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Halo Scarab — Artifact Creature — Insect {2}, 2/1 (Streets of New
// Capenna, #5491):
//
//	"{2}, Exile this card from your graveyard: Create a Treasure token."
//
// Goldmeadow Nomad's graveyard activation: exiling the card is part of the
// cost (CR 602.2b), and it has no timing restriction.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1201c146-54af-4668-a544-cde3ece23f2e",
		Name:         "Halo Scarab",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}, Exile this card from your graveyard: Create a Treasure token.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Tokens: 1},
			Cost:    Plus(ManaCost("{2}"), ExileThis()),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect:  Do(CreateToken{Template: TreasureToken(), N: 1}),
		}},
	})
}
