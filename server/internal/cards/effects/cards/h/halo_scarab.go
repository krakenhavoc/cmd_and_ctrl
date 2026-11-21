package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "1201c146-54af-4668-a544-cde3ece23f2e",
		Name:         "Halo Scarab",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, Exile this card from your graveyard: Create a Treasure token.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Tokens: 1},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.ExileThis()),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect:  effects.Do(effects.CreateToken{Template: effects.TreasureToken(), N: 1}),
		}},
	})
}
